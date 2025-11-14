package service

import (
	chatdomain "chatsrv/internal/domain/chat"
	msgdomain "chatsrv/internal/domain/msg"
	"context"

	"golang.org/x/net/websocket"
)

type ChatService interface {
	GetIncomeMessage(ws *websocket.Conn, msg msgdomain.Message) error
	GetChats(ctx context.Context) ([]*chatdomain.Chat, error)
	HandleDisconnect(ws *websocket.Conn, clientID string)
}
