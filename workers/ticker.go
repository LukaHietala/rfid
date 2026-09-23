package workers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/lukahietala/rfid/db"
	"github.com/lukahietala/rfid/websockets"
)

func StartDoneTicker(ctx context.Context, store *db.Store, hub *websockets.Hub, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := store.IncrementDoneSeconds(ctx, int(interval.Seconds())); err != nil {
				log.Println("failed to increment done_seconds:", err)
				continue
			}

			students, err := store.ListStudents(ctx)
			if err != nil {
				log.Println("error listing students:", err)
				continue
			}

			bytes, err := json.Marshal(students)
			if err != nil {
				log.Println("error marshaling students:", err)
				continue
			}

			hub.Broadcast(websockets.Event{
				Event:   "students:update",
				Payload: bytes,
			})
		}
	}
}
