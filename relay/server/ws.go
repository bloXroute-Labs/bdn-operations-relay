package server

import (
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
