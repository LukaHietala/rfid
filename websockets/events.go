package websockets

import (
	"encoding/json"
)

type Event struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}
