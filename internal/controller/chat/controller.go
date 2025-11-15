package chatctrl

import (
	"chatsrv/internal/controller"
	chatdomain "chatsrv/internal/domain/chat"
	msgdomain "chatsrv/internal/domain/msg"
	"chatsrv/internal/service"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

var _ controller.ChatController = (*implementation)(nil)

type Option func(*implementation)

func WithLogger(log *zap.Logger) Option {
	return func(i *implementation) {
		i.log = log
	}
}

func WithService(srv service.ChatService) Option {
	return func(i *implementation) {
		i.srv = srv
	}
}

func NewChatController(opts ...Option) controller.ChatController {
	impl := &implementation{}

	for _, opt := range opts {
		opt(impl)
	}

	return impl
}

type implementation struct {
	log *zap.Logger
	srv service.ChatService
}

// CreateChat implements controller.ChatController.
func (c *implementation) CreateChat(w http.ResponseWriter, r *http.Request) {
	var req chatdomain.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.log.Error("failed to decode request", zap.Error(err))
		http.Error(w, "failed to decode request", http.StatusBadRequest)
		return
	}

	chat, err := c.srv.CreateChat(r.Context(), req.Name)
	if err != nil {
		c.log.Error("failed to create chat", zap.Error(err))
		http.Error(w, "failed to create chat", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(chat)
	if err != nil {
		c.log.Error("failed to marshal chat", zap.Error(err))
		http.Error(w, "failed to marshal chat", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// GetChats implements controller.ChatController.
func (c *implementation) GetChats(w http.ResponseWriter, r *http.Request) {
	chats, err := c.srv.GetChats(r.Context())
	if err != nil {
		c.log.Error("failed to get chats", zap.Error(err))
		http.Error(w, "failed to get chats", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(chats)
	if err != nil {
		c.log.Error("failed to marshal chats", zap.Error(err))
		http.Error(w, "failed to marshal chats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (c *implementation) HandleWebSocket(ws *websocket.Conn) {
	var client msgdomain.Message

	defer func() {
		if client.SenderID != "" {
			c.srv.HandleDisconnect(ws, client.SenderID)

			msg := msgdomain.Message{
				Content:  fmt.Sprintf("%s left chat", client.SenderID),
				SenderID: "SYSTEM",
				ChatID:   client.ChatID,
			}
			c.srv.GetIncomeMessage(ws, msg)
		}
		err := ws.Close()
		if err != nil {
			c.log.Error("error close websocket connection", zap.Error(err))
		}
	}()

	c.log.Info("WebSocket client connected", zap.String("local", ws.LocalAddr().String()))

	var msg msgdomain.Message
	for {
		select {
		case <-ws.Request().Context().Done():
			c.log.Info("WebSocket client context done")
			return
		default:
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				c.log.Info("WebSocket client disconnected", zap.Error(err))
				return
			}
			client = msg
			err = c.srv.GetIncomeMessage(ws, msg)
			if err != nil {
				c.log.Error("error getting income message", zap.Error(err))
				continue
			}
			c.log.Debug("received message", zap.Any("msg", msg))
		}
	}
}
