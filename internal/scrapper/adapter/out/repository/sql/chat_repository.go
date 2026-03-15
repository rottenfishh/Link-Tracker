package sql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) SaveChat(ctx context.Context, chat *model.Chat) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO chats(id)
     VALUES($1)
     ON CONFLICT (id) DO NOTHING`,
		chat.ChatID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *ChatRepository) GetChats(ctx context.Context) ([]model.Chat, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM chats")
	if err != nil {
		return nil, err
	}
	var chats []model.Chat
	for rows.Next() {
		var chat model.Chat
		err = rows.Scan(&chat.ChatID, &chat.UserID)
		if err != nil {
			return nil, err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
	}
	return chats, nil
}

func (r *ChatRepository) DeleteChat(ctx context.Context, chatId int64) (*model.Chat, error) {
	row := r.db.QueryRowContext(ctx, "DELETE FROM chats WHERE id = $1 RETURNING id", chatId)

	var chat model.Chat
	err := row.Scan(&chat.ChatID, &chat.UserID)
	if err != nil {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	return &chat, nil
}
