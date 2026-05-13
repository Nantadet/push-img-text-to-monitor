package ws

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

var Clients = make(map[*websocket.Conn]bool)

var Mutex sync.Mutex
