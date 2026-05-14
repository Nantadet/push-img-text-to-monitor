package ws

import (
	"encoding/json"
	"sync"

	"github.com/fasthttp/websocket"
)

var Clients = make(map[*websocket.Conn]bool)

var Mutex sync.Mutex

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

func Broadcast(eventType string, data any) {
	payload, err := json.Marshal(Event{
		Type: eventType,
		Data: data,
	})
	if err != nil {
		return
	}

	Mutex.Lock()
	defer Mutex.Unlock()

	for conn := range Clients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			_ = conn.Close()
			delete(Clients, conn)
		}
	}
}
