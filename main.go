package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/lukahietala/rfid/api"
	"github.com/lukahietala/rfid/db"
	"github.com/lukahietala/rfid/websockets"
)

func main() {
	conn, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	store := db.NewStore(conn)

	hub := websockets.NewHub()
	go hub.Run()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go accumulateDoneTicker(ctx, store, hub)

	ln, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println("accept err", err)
				continue
			}
			go handleReadConn(conn, store, hub)

		}
	}()

	r := api.NewRouter(store, hub)

	log.Println("Server running on :3000")
	http.ListenAndServe(":3000", r)
}

func accumulateDoneTicker(ctx context.Context, store *db.Store, hub *websockets.Hub) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := store.IncrementDoneSeconds(ctx, 10); err != nil {
				log.Println("failed to increment done_seconds:", err)
				continue
			}

			students, err := store.ListStudents(ctx)
			if err != nil {
				log.Println("error:", err)
			}
			bytes, err := json.Marshal(students)
			if err != nil {
				log.Println("error:", err)
			}

			hub.Broadcast(websockets.Event{
				Event:   "students:update",
				Payload: bytes,
			})
		}
	}
}

func handleReadConn(conn net.Conn, store *db.Store, hub *websockets.Hub) {
	defer conn.Close()
	for {
		var header [48]byte
		if _, err := io.ReadFull(conn, header[:]); err != nil {
			log.Println(err)
			break
		}

		magic := header[0]
		if magic != 0xAA {
			log.Println("invalid magic byte")
			break
		}

		uidLen := header[1]
		if uidLen != 4 && uidLen != 7 {
			log.Printf("invalid payload length of %d. Expected 4 or 7", uidLen)
			break
		}

		timestamp := int64(binary.BigEndian.Uint64(header[2:10]))
		now := time.Now().Unix()
		if timestamp < now-5 || timestamp > now+5 {
			log.Println("clock skew is not in valid range")
			break
		}

		// TODO: check seq
		seq := header[10:14]

		readerID := binary.BigEndian.Uint16(header[14:16])

		devCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		dev, err := store.FindDeviceByID(devCtx, int(readerID))
		if err != nil {
			log.Printf("reader %d not in allowed devices", readerID)
			break
		}

		uid := make([]byte, uidLen)
		if _, err := io.ReadFull(conn, uid); err != nil {
			log.Println(err)
			break
		}

		devHash := header[16:48]
		mac := hmac.New(sha256.New, []byte(dev.SecretKey))

		_ = binary.Write(mac, binary.BigEndian, magic)
		_ = binary.Write(mac, binary.BigEndian, uidLen)
		_ = binary.Write(mac, binary.BigEndian, timestamp)
		_ = binary.Write(mac, binary.BigEndian, seq)
		_ = binary.Write(mac, binary.BigEndian, readerID)

		mac.Write(uid)

		expectedHash := mac.Sum(nil)
		if !hmac.Equal(expectedHash, devHash) {
			log.Println("invalid hash, check secret key and reader id")
			break
		}

		uidStr := hex.EncodeToString(uid)

		studentGetCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		student, err := store.FindStudentByUID(studentGetCtx, uidStr)
		if err != nil {
			log.Println(err)
			continue
		}

		studentUpdateCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if student.Status == "IN" {
			student.Status = "OUT"
		} else {
			student.Status = "IN"
		}

		err = store.UpdateStudent(studentUpdateCtx, student.ID, *student)
		if err != nil {
			log.Println("failed to update student:", err)
			continue
		}

		studentBytes, err := json.Marshal(student)
		if err != nil {
			log.Println(err)
			continue
		}

		hub.Broadcast(websockets.Event{
			Event:   "student:update",
			Payload: studentBytes,
		})

		scanCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		scan := &db.Scan{
			UID:       uidStr,
			StudentID: student.ID,
			Timestamp: time.Unix(timestamp, 0).Format(time.DateTime),
		}
		err = store.NewScan(scanCtx, scan)
		if err != nil {
			log.Printf("failed to add scan %s", uidStr)
			continue
		}

		scanBytes, err := json.Marshal(scan)
		if err != nil {
			log.Println(err)
			continue
		}

		hub.Broadcast(websockets.Event{
			Event:   "scan:new",
			Payload: scanBytes,
		})

	}
}
