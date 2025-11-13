package routes

import (
	"chatsrv/internal/controller"
	"net/http"
	"net/url"

	"golang.org/x/net/websocket"
)

func InitRoutes(ctrl controller.ChatController) *http.ServeMux {
	mux := http.NewServeMux()

	wsServer := &websocket.Server{
		Handler: func(ws *websocket.Conn) {
			ctrl.HandleWebSocket(ws)
		},
		Handshake: func(config *websocket.Config, r *http.Request) error {
			originStr := r.Header.Get("Origin")
			if originStr != "" {
				if origin, err := url.Parse(originStr); err == nil {
					config.Origin = origin
				}
			}
			return nil
		},
	}

	mux.Handle("/ws", wsServer)
	mux.HandleFunc("/chats", ctrl.GetChats)

	return mux
}
