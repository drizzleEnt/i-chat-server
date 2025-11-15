package chatsrv

import (
	chatdomain "chatsrv/internal/domain/chat"
	msgdomain "chatsrv/internal/domain/msg"
	"chatsrv/internal/repository"
	"chatsrv/internal/service"
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

var _ service.ChatService = (*chatService)(nil)

func NewChatService(repo repository.ChatRepository, log *zap.Logger) service.ChatService {
	s := &chatService{
		chats:   make(map[string]*chat),
		msgChan: make(chan msgdomain.Message, 100),
		repo:    repo,
		log:     log,
	}

	go s.processMessage()

	return s
}

type chatService struct {
	mutex sync.RWMutex
	chats map[string]*chat

	msgChan chan msgdomain.Message
	repo    repository.ChatRepository
	log     *zap.Logger
}

// CreateChat implements service.ChatService.
func (s *chatService) CreateChat(ctx context.Context, name string) (*chatdomain.Chat, error) {
	uuid := uuid.New().String()

	err := s.repo.CreateChat(ctx, uuid, name)
	if err != nil {
		s.log.Error("CreateChat",
			zap.Any("msg", name),
			zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func (s *chatService) HandleDisconnect(ws *websocket.Conn, clientID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, ch := range s.chats {
		if _, ok := ch.clients[clientID]; ok {
			ch.removeClient(clientID)
			s.log.Debug("Client removed from chat on disconnect",
				zap.Any("Client", clientID),
				zap.Any("Chat", ch.chatID))
		}
	}
}

// GetChats implements service.ChatService.
func (c *chatService) GetChats(ctx context.Context) ([]*chatdomain.Chat, error) {
	return c.repo.GetChats(ctx)
}

func (c *chatService) GetIncomeMessage(ws *websocket.Conn, msg msgdomain.Message) error {
	switch msg.Action {
	case string(msgdomain.ActionJoinChat):
		c.log.Debug("Handle Join Chat",
			zap.Any("User", msg.SenderID),
			zap.Any("Chat", msg.ChatID))
		return c.handleJoinChat(ws, msg)
	case string(msgdomain.ActionLeaveChat):
		c.log.Debug("Handle Leave Chat",
			zap.Any("User", msg.SenderID),
			zap.Any("Chat", msg.ChatID))
		return c.handleLeaveChat(ws, msg)
	case string(msgdomain.ActionSendText), string(msgdomain.ActionSendBinary):
		// For simplicity, we treat both text and binary messages the same way he
		c.log.Debug("Handle Send Text",
			zap.Any("User", msg.SenderID),
			zap.Any("Chat", msg.ChatID))
		c.msgChan <- msg
		return nil
	default:
		// Unknown action
		return nil
	}

	return nil
}

func (c *chatService) handleJoinChat(ws *websocket.Conn, msg msgdomain.Message) error {
	c.mutex.RLock()
	chat, ok := c.chats[msg.ChatID]
	c.mutex.RUnlock()
	if !ok {
		storedChat, err := c.repo.GetChat(ws.Request().Context(), msg.ChatID)
		if err != nil {
			c.log.Debug("Join Chat",
				zap.Any("msg", msg),
				zap.Error(err))
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
		c.log.Error("Join Chat already in chat",
			zap.Any("user", msg.SenderID),
			zap.Any("chat", msg.ChatID))
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
		c.log.Error("Leave Chat with unknown chatID",
			zap.Any("user", msg.SenderID),
			zap.Any("chat", msg.ChatID))
		return fmt.Errorf("unknown chatID %s", msg.ChatID)
	}

	chat.m.RLock()
	client, ok := chat.clients[msg.SenderID]
	chat.m.RUnlock()
	if !ok {
		c.log.Error("Leave Chat user not found in chat",
			zap.Any("user", msg.SenderID),
			zap.Any("chat", msg.ChatID))
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
				if c.id == msg.SenderID {
					continue
				}
				clients = append(clients, c)
			}
			chat.m.RUnlock()

			// Broadcast the message to all clients in the chat
			for _, client := range clients {
				err := client.sendMessage(msg)
				if err != nil {
					c.log.Error("processMessage",
						zap.Any("msg", msg),
						zap.Any("client", client.id),
						zap.Any("chat", msg.ChatID),
						zap.Error(err))
					continue
				}
			}
		default:
			// Process the message (e.g., broadcast to other clients)
		}
	}
}
