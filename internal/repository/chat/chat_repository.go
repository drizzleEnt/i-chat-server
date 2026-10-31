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

// NextChatID implements repository.ChatRepository.
func (c *chatRepository) NextChatID(ctx context.Context, chatType int) (int64, error) {
	query := `
	UPDATE id_sequences
	SET last_value = last_value + 1
	WHERE name = 'chat_id'
	RETURNING last_value
	`
	var lastValue int64
	err := c.db.QueryRowContext(ctx, query).Scan(&lastValue)
	if err != nil {
		return 0, err
	}

	localID := uint64(lastValue) & 0x3FFFFFFFFFFFFFFF

	return int64((uint64(chatType) << 62) | localID), nil
}

// CreateChat implements repository.ChatRepository.
func (c *chatRepository) CreateChat(ctx context.Context, chatType int, name string) (int64, error) {
	chatID, err := c.NextChatID(ctx, chatType)
	if err != nil {
		return 0, err
	}

	query := `
	INSERT INTO 
	chats(id, name) 
	VALUES ($1, $2)
	ON CONFLICT DO NOTHING`

	if _, err := c.db.ExecContext(ctx, query, chatID, name); err != nil {
		return 0, err
	}

	return chatID, nil
}

// GetChat implements repository.ChatRepository.
func (c *chatRepository) GetChat(ctx context.Context, chatID int64) (*chatdomain.Chat, error) {
	query := `SELECT id, name FROM chats WHERE id = $1`

	var chat chatdomain.Chat
	err := c.db.QueryRowContext(ctx, query, chatID).Scan(&chat.ID, &chat.Name)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

// GetChats implements repository.ChatRepository.
func (c *chatRepository) GetChats(ctx context.Context) ([]*chatdomain.Chat, error) {
	query := "SELECT id, name FROM chats"
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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
