package service

import (
	chatdomain "chatsrv/internal/domain/chat"
	"chatsrv/internal/domain/msg"
	"context"

	"golang.org/x/net/websocket"
)

type ChatService interface {
	GetIncomeMessage(ws *websocket.Conn, msg msg.Message) error
	GetChats(ctx context.Context) ([]*chatdomain.Chat, error)
}
