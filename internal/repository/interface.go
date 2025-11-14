package repository

import (
	chatdomain "chatsrv/internal/domain/chat"
	"context"
)

type ChatRepository interface {
	GetChats(ctx context.Context) ([]*chatdomain.Chat, error)
	GetChat(ctx context.Context, chatID string) (*chatdomain.Chat, error)
	CreateChat(ctx context.Context, chatID string) error
}
