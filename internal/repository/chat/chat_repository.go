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

// GetChat implements repository.ChatRepository.
func (c *chatRepository) GetChat(ctx context.Context, chatID string) (*chatdomain.Chat, error) {
	query := `SELECT uuid, name FROM chats WHERE uuid = $1`

	var chat chatdomain.Chat
	err := c.db.QueryRowContext(ctx, query, chatID).Scan(&chat.ID, &chat.Name)
	if err != nil {
		return nil, err
	}

	return &chat, nil
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
