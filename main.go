package main

import (
	"context"
	"encoding/json"
	"log"
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
