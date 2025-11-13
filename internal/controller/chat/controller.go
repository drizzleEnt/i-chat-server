package chatctrl

import (
	"chatsrv/internal/controller"
	"chatsrv/internal/domain/msg"
	"chatsrv/internal/service"
	"encoding/json"
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
	defer func() {
		err := ws.Close()
		if err != nil {
			c.log.Error("error close websocket connection", zap.Error(err))
		}
	}()

	c.log.Info("WebSocket client connected", zap.String("local", ws.LocalAddr().String()))

	var msg msg.Message
	for {
		select {
		case <-ws.Request().Context().Done():
			c.log.Info("WebSocket client context done")
			return
		default:
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				c.log.Info("WebSocket client disconnected", zap.Error(err))
				break
			}
			err = c.srv.GetIncomeMessage(ws, msg)
			if err != nil {
				c.log.Error("error getting income message", zap.Error(err))
				continue
			}
			c.log.Debug("received message", zap.Any("msg", msg))
		}
	}
}
