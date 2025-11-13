package repository

import (
	chatdomain "chatsrv/internal/domain/chat"
	"context"
)

type ChatRepository interface {
	GetChats(ctx context.Context) ([]*chatdomain.Chat, error)
}
