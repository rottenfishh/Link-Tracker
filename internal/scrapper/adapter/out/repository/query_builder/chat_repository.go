package query_builder

import (
    "context"
    "database/sql"
    "errors"

    sq "github.com/Masterminds/squirrel"
    "github.com/jackc/pgx/v5"
    "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type ChatRepository struct {
	db *sql.DB
    psql sq.StatementBuilderType
}

func NewChatRepository(db *sql.DB, psql sq.StatementBuilderType) *ChatRepository {
    return &ChatRepository{db:db, psql:psql}
}

// Suffix("RETURNING id").
func (r *ChatRepository) SaveChat(ctx context.Context, chat *model.Chat) error {
	_, err := r.psql.Insert("chats").Columns("id").Values(chat.ChatID).
        RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

//TODO: как избавиться от дублирования кода запроса к бд?
func (r *ChatRepository) GetChats(ctx context.Context) ([]model.Chat, error) {
	rows, err := r.psql.Select("*").From("chats").RunWith(r.db).QueryContext(ctx)
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

//TODO: check error what is it
func (r *ChatRepository) DeleteChat(ctx context.Context, chatId int64) (*model.Chat, error) {
    var chat model.Chat
	err := r.psql.Delete("chats").Where(sq.Eq{"id": chatId}).
        Suffix("RETURNING id").RunWith(r.db).QueryRowContext(ctx).Scan(&chat.ChatID, &chat.UserID)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, model.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &chat, nil
}
