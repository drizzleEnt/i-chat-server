package chatsrv

import (
	"chatsrv/internal/domain/msg"

	"golang.org/x/net/websocket"
)

func NewClient(id string, chatID string, conn *websocket.Conn) *client {
	return &client{
		id:     id,
		chatID: chatID,
		conn:   conn,
	}
}

type client struct {
	id     string
	chatID string
	conn   *websocket.Conn
}

func (c *client) sendMessage(message msg.Message) error {
	msg := msg.Message{
		Content:  message.Content,
		SenderID: message.SenderID,
		ChatID:   message.ChatID,
	}
	return websocket.JSON.Send(c.conn, msg)
}
