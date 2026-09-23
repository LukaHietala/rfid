package tcp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/lukahietala/rfid/db"
	"github.com/lukahietala/rfid/websockets"
)

const (
	MagicByte  byte  = 0xAA
	HeaderSize int   = 48
	MaxSkew    int64 = 5 // seconds
)

type Server struct {
	addr  string
	store *db.Store
	hub   *websockets.Hub
	ln    net.Listener
}

func New(addr string, store *db.Store, hub *websockets.Hub) *Server {
	return &Server{
		addr:  addr,
		store: store,
		hub:   hub,
	}
}

func (s *Server) Start(ctx context.Context) error {
	var err error
	s.ln, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer s.ln.Close()

	log.Printf("rfid tcp server listening on %s", s.addr)

	go func() {
		<-ctx.Done()
		s.ln.Close()
	}()

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Println("accept err:", err)
				continue
			}
		}
		go s.handleConn(conn)
	}
}

func (s *Server) Close() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	for {
		if err := s.processPacket(conn); err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println(err)
			}
			break
		}
	}
}

func (s *Server) processPacket(conn net.Conn) error {
	var header [HeaderSize]byte
	if _, err := io.ReadFull(conn, header[:]); err != nil {
		return err
	}

	magic := header[0]
	if magic != MagicByte {
		return errors.New("invalid magic byte")
	}

	uidLen := header[1]
	if uidLen != 4 && uidLen != 7 {
		return fmt.Errorf("invalid payload length %d (expected 4 or 7)", uidLen)
	}

	timestamp := int64(binary.BigEndian.Uint64(header[2:10]))
	now := time.Now().Unix()
	if timestamp < now-MaxSkew || timestamp > now+MaxSkew {
		// Could be a replay attack
		return errors.New("clock skew out of valid range")
	}

	// TODO: Track seq
	seq := header[10:14]
	readerID := binary.BigEndian.Uint16(header[14:16])

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dev, err := s.store.FindDeviceByID(ctx, int(readerID))
	if err != nil {
		return fmt.Errorf("reader %d not in allowed devices: %w", readerID, err)
	}

	uid := make([]byte, uidLen)
	if _, err := io.ReadFull(conn, uid); err != nil {
		return fmt.Errorf("failed to read uid payload: %w", err)
	}

	devHash := header[16:48]
	mac := hmac.New(sha256.New, []byte(dev.SecretKey))

	_ = binary.Write(mac, binary.BigEndian, magic)
	_ = binary.Write(mac, binary.BigEndian, uidLen)
	_ = binary.Write(mac, binary.BigEndian, timestamp)
	_ = binary.Write(mac, binary.BigEndian, seq)
	_ = binary.Write(mac, binary.BigEndian, readerID)
	mac.Write(uid)

	if !hmac.Equal(mac.Sum(nil), devHash) {
		return errors.New("invalid hmac hash")
	}

	return s.handleScan(ctx, hex.EncodeToString(uid), timestamp)
}

func (s *Server) handleScan(ctx context.Context, uidStr string, timestamp int64) error {
	student, err := s.store.FindStudentByUID(ctx, uidStr)
	if err != nil {
		return fmt.Errorf("student not found for uid %s: %w", uidStr, err)
	}

	if student.Status == "IN" {
		student.Status = "OUT"
	} else {
		student.Status = "IN"
	}

	if err := s.store.UpdateStudent(ctx, student.ID, *student); err != nil {
		return err
	}

	if bytes, err := json.Marshal(student); err == nil {
		s.hub.Broadcast(websockets.Event{
			Event:   "student:update",
			Payload: bytes,
		})
	}

	scan := &db.Scan{
		UID:       uidStr,
		StudentID: student.ID,
		Timestamp: time.Unix(timestamp, 0).Format(time.DateTime),
	}
	if err := s.store.NewScan(ctx, scan); err != nil {
		return fmt.Errorf("failed to create scan: %w", err)
	}

	if bytes, err := json.Marshal(scan); err == nil {
		s.hub.Broadcast(websockets.Event{
			Event:   "scan:new",
			Payload: bytes,
		})
	}

	return nil
}
