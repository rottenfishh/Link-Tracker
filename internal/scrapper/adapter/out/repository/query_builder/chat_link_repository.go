package query_builder

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type ChatLinkRepository struct {
	db   *sql.DB
	psql sq.StatementBuilderType
}

func NewChatLinkRepository(db *sql.DB, psql sq.StatementBuilderType) *ChatLinkRepository {
	return &ChatLinkRepository{db: db, psql: psql}
}

func (r *ChatLinkRepository) GetLinksByChatID(ctx context.Context, chatId int64) ([]model.Link, error) {
	rows, err := r.psql.Select("l.id, l.link, l.domain, l.last_updated").From("links l").
		Join("chat_link cl ON cl.link_id = l.id").Where(sq.Eq{"cl.chat_id": chatId}).
		RunWith(r.db).QueryContext(ctx)
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
	_, err := r.psql.Insert("chat_link").Columns("chat_id", "link_id").
		Values(chatId, linkId).RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *ChatLinkRepository) Unsubscribe(ctx context.Context, chatId int64, linkId int64) error {
	_, err := r.psql.Delete("chat_link").Where(sq.Eq{"chat_id": chatId, "link_id": linkId}).
		RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *ChatLinkRepository) GetChatsByLinkID(ctx context.Context, linkID int64) ([]model.Chat, error) {
	rows, err := r.psql.Select("*").From("chats c").
		Join("chat_link cl on cl.chat_id = c.id").
		Where(sq.Eq{"cl.link_id": linkID}).RunWith(r.db).QueryContext(ctx)
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
	rows, err := r.psql.Select("chat_id").From("chat_link").Where(sq.Eq{"link_id": linkId}).
		RunWith(r.db).QueryContext(ctx)
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
