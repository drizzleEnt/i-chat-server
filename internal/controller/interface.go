package controller

import (
	"net/http"

	"golang.org/x/net/websocket"
)

type ChatController interface {
	HandleWebSocket(ws *websocket.Conn)
	GetChats(w http.ResponseWriter, r *http.Request)
	CreateChat(w http.ResponseWriter, r *http.Request)
}
