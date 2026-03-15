package sql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type ChatLinkRepository struct {
	db *pgxpool.Pool
}

func NewChatLinkRepository(db *pgxpool.Pool) *ChatLinkRepository {
	return &ChatLinkRepository{db: db}
}

func (r *ChatLinkRepository) GetLinksByChatID(ctx context.Context, chatId int64) ([]model.Link, error) {
	sql := `SELECT l.id, l.link, l.domain, l.last_updated
            FROM links l
            JOIN chat_link cl ON cl.link_id = l.id
            WHERE cl.chat_id = $1`
	rows, err := r.db.Query(ctx, sql, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]model.Link, 0)
	for rows.Next() {
		var link model.Link
		err = rows.Scan(&link.Id, &link.Link, &link.Domain, &link.LastUpdated)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *ChatLinkRepository) Subscribe(ctx context.Context, chatId int64, linkId int64) error {
	sql := `INSERT INTO chat_link(chat_id, link_id) VALUES ($1, $2)`
	_, err := r.db.Exec(ctx, sql, chatId, linkId)
	if err != nil {
		return err
	}
	return nil
}

func (r *ChatLinkRepository) Unsubscribe(ctx context.Context, chatId int64, linkId int64) error {
	sql := "DELETE FROM chat_link WHERE chat_id = $1 AND link_id = $2"
	_, err := r.db.Exec(ctx, sql, chatId, linkId)
	if err != nil {
		return err
	}
	return nil
}

func (r *ChatLinkRepository) GetChatsByLinkID(ctx context.Context, linkID int64) ([]model.Chat, error) {
	sql := `SELECT c.id, c.user_id
            FROM chats c
            JOIN chat_link cl ON cl.chat_id = c.id
            WHERE cl.link_id = $1`
	rows, err := r.db.Query(ctx, sql, linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	chats := make([]model.Chat, 0)
	for rows.Next() {
		var c model.Chat
		err = rows.Scan(&c.ChatID, &c.UserID)
		if err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *ChatLinkRepository) GetChatIdsByLinkID(ctx context.Context, linkId int64) ([]int64, error) {
	sql := "SELECT chat_id FROM chat_link WHERE link_id = $1"
	rows, err := r.db.Query(ctx, sql, linkId)
	if err != nil {
		return nil, err
	}
	chatIds := make([]int64, 0)
	for rows.Next() {
		var chatId int64
		err = rows.Scan(&chatId)
		if err != nil {
			return nil, err
		}
		chatIds = append(chatIds, chatId)
	}
	if err = rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return chatIds, nil
}
