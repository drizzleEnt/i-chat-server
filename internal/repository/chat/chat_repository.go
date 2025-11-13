package chatrepository

import (
	chatdomain "chatsrv/internal/domain/chat"
	"chatsrv/internal/repository"
	"context"
	"database/sql"
)

var _ repository.ChatRepository = (*chatRepository)(nil)

func NewChatRepository(db *sql.DB) *chatRepository {
	return &chatRepository{
		db: db,
	}
}

type chatRepository struct {
	db *sql.DB
}

// GetChats implements repository.ChatRepository.
func (c *chatRepository) GetChats(ctx context.Context) ([]*chatdomain.Chat, error) {
	query := "SELECT uuid, name FROM chats"
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*chatdomain.Chat
	for rows.Next() {
		var chat chatdomain.Chat
		if err := rows.Scan(&chat.ID, &chat.Name); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}
