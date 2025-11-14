package chatsrv

import (
	chatdomain "chatsrv/internal/domain/chat"
	msgdomain "chatsrv/internal/domain/msg"
	"chatsrv/internal/repository"
	"chatsrv/internal/service"
	"context"
	"fmt"
	"log"
	"sync"

	"golang.org/x/net/websocket"
)

var _ service.ChatService = (*chatService)(nil)

func NewChatService(repo repository.ChatRepository) service.ChatService {
	s := &chatService{
		chats:   make(map[string]*chat),
		msgChan: make(chan msgdomain.Message, 100),
		repo:    repo,
	}

	go s.processMessage()

	return s
}

type chatService struct {
	mutex sync.RWMutex
	chats map[string]*chat

	msgChan chan msgdomain.Message
	repo    repository.ChatRepository
}

// GetChats implements service.ChatService.
func (c *chatService) GetChats(ctx context.Context) ([]*chatdomain.Chat, error) {
	return c.repo.GetChats(ctx)
}

func (c *chatService) GetIncomeMessage(ws *websocket.Conn, msg msgdomain.Message) error {
	switch msg.Action {
	case string(msgdomain.ActionJoinChat):
		log.Printf("Handle Join Chat %s user-%s", msg.ChatID, msg.SenderID)
		return c.handleJoinChat(ws, msg)
	case string(msgdomain.ActionLeaveChat):
		log.Printf("Handle Leave Chat %s user-%s", msg.ChatID, msg.SenderID)
		return c.handleLeaveChat(ws, msg)
	case string(msgdomain.ActionSendText), string(msgdomain.ActionSendBinary):
		// For simplicity, we treat both text and binary messages the same way he
		log.Printf("Handle Send Text to Chat %s user-%s", msg.ChatID, msg.SenderID)
		c.msgChan <- msg
		return nil
	case string(msgdomain.ActionCreateChat):
		log.Printf("Handle Create Chat %s user-%s", msg.ChatID, msg.SenderID)
		return c.handleCreateChat(msg)
	default:
		// Unknown action
		return nil
	}

	return nil
}

func (c *chatService) handleCreateChat(msg msgdomain.Message) error {
	c.mutex.RLock()
	_, ok := c.chats[msg.ChatID]
	c.mutex.RUnlock()
	if ok {
		log.Printf("CreateChat get msg with exist chatID %+v", msg)
		return fmt.Errorf("already exist chatID %s", msg.ChatID)
	}

	chat := newChat(msg.ChatID)
	c.mutex.Lock()
	c.chats[msg.ChatID] = chat
	c.mutex.Unlock()

	return nil
}

func (c *chatService) handleJoinChat(ws *websocket.Conn, msg msgdomain.Message) error {
	c.mutex.RLock()
	chat, ok := c.chats[msg.ChatID]
	c.mutex.RUnlock()
	if !ok {
		storedChat, err := c.repo.GetChat(ws.Request().Context(), msg.ChatID)
		if err != nil {
			log.Printf("JoinChat failed get chats %+v", msg)
			return err
		}
		chat = newChat(storedChat.ID)
		c.mutex.Lock()
		c.chats[msg.ChatID] = chat
		c.mutex.Unlock()
	}

	chat.m.RLock()
	_, ok = chat.clients[msg.SenderID]
	chat.m.RUnlock()
	if ok {
		log.Printf("JoinChat user already in chat %+v", msg)
		return fmt.Errorf("user %s already in chat %s", msg.SenderID, msg.ChatID)
	}

	chat.addClient(NewClient(msg.SenderID, msg.ChatID, ws))
	return nil
}

func (c *chatService) handleLeaveChat(ws *websocket.Conn, msg msgdomain.Message) error {
	c.mutex.RLock()
	chat, ok := c.chats[msg.ChatID]
	c.mutex.RUnlock()
	if !ok {
		log.Printf("LeaveChat get msg with unknown chatID %+v", msg)
		return fmt.Errorf("unknown chatID %s", msg.ChatID)
	}

	chat.m.RLock()
	client, ok := chat.clients[msg.SenderID]
	chat.m.RUnlock()
	if !ok {
		log.Printf("LeaveChat user not found in chat %+v", msg)
		return fmt.Errorf("user %s not found in chat %s", msg.SenderID, msg.ChatID)
	}

	chat.m.Lock()
	delete(chat.clients, client.id)
	chat.m.Unlock()
	if len(chat.clients) == 0 {
		c.mutex.Lock()
		delete(c.chats, msg.ChatID)
		c.mutex.Unlock()
	}
	return nil
}

func (c *chatService) handleIncomingMessage(ws *websocket.Conn, msg msgdomain.Message) error {
	c.mutex.RLock()
	chat, ok := c.chats[msg.ChatID]
	c.mutex.RUnlock()
	if !ok {
		chat = newChat(msg.ChatID)
		c.mutex.Lock()
		c.chats[msg.ChatID] = chat
		c.mutex.Unlock()
	}

	chat.m.RLock()
	client, ok := chat.clients[msg.SenderID]
	chat.m.RUnlock()
	if !ok {
		client = NewClient(msg.SenderID, msg.ChatID, ws)
		chat.addClient(client)
	}

	c.msgChan <- msg
	return nil
}

func (c *chatService) processMessage() error {
	for {
		select {
		case msg := <-c.msgChan:
			c.mutex.RLock()
			chat, ok := c.chats[msg.ChatID]
			c.mutex.RUnlock()
			if !ok {
				// Handle error: chat not found

				continue
			}

			chat.m.RLock()
			clients := make([]*client, 0, len(chat.clients))
			for _, c := range chat.clients {
				clients = append(clients, c)
			}
			chat.m.RUnlock()

			// Broadcast the message to all clients in the chat
			for _, client := range clients {
				err := client.sendMessage(msg)
				if err != nil {
					// Handle error sending message to client
					continue
				}
			}
		// Handle the incoming message
		// For example, you can route it to the appropriate chat room
		// based on msg.ChatID
		// Here, we just print it for demonstration purposes
		// log.Printf("Received message: %+v", msg)
		// You can add more logic here to manage chats and clients
		default:
			// Process the message (e.g., broadcast to other clients)
		}
	}
}

func (c *chatService) sendMsg(chatID int64) error {

	return nil
}
