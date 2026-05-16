//nolint:wrapcheck,funcorder // repository methods mostly proxy storage errors and keep generated-like method grouping
package querybuilder

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
)

type ChatLinkRepository struct {
	db   *sql.DB
	psql sq.StatementBuilderType
}

func (r *ChatLinkRepository) GetLinksByChatIDAndTag(ctx context.Context, chatID int64, tag string) ([]model.Link, error) {
	rows, err := r.psql.Select("l.ID, l.link, l.domain, l.last_updated").From("links l").
		Join("chat_link cl on cl.link_ID = l.ID").Join("link_tag lt on lt.link_ID = l.ID").
		Join("tags t on t.ID = lt.tag_ID").
		Where(sq.Eq{"cl.chat_ID": chatID}).Where(sq.Eq{"t.name": tag}).RunWith(r.db).QueryContext(ctx)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing query builder chat links rows", "error", closeErr)
		}
	}()

	links := make([]model.Link, 0)
	for rows.Next() {
		var link model.Link
		if err = rows.Scan(&link.ID, &link.Link, &link.Domain, &link.LastUpdated); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	slog.Info("repo links", "links", links)
	return links, nil
}

func NewChatLinkRepository(db *sql.DB, psql sq.StatementBuilderType) *ChatLinkRepository {
	return &ChatLinkRepository{db: db, psql: psql}
}

func (r *ChatLinkRepository) GetLinksByChatID(ctx context.Context, chatID int64) ([]model.Link, error) {
	rows, err := r.psql.Select("l.ID, l.link, l.domain, l.last_updated").From("links l").
		Join("chat_link cl ON cl.link_ID = l.ID").Where(sq.Eq{"cl.chat_ID": chatID}).
		RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing query builder chat ID rows", "error", closeErr)
		}
	}()

	links := make([]model.Link, 0)
	for rows.Next() {
		var link model.Link
		err = rows.Scan(&link.ID, &link.Link, &link.Domain, &link.LastUpdated)
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

func (r *ChatLinkRepository) Subscribe(ctx context.Context, chatID int64, linkID int64) error {
	exec := repository.GetExecutor(ctx, r.db)

	_, err := r.psql.Insert("chat_link").Columns("chat_ID", "link_ID").
		Values(chatID, linkID).RunWith(exec).ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *ChatLinkRepository) Unsubscribe(ctx context.Context, chatID int64, linkID int64) error {
	exec := repository.GetExecutor(ctx, r.db)

	_, err := r.psql.Delete("chat_link").Where(sq.Eq{"chat_ID": chatID, "link_ID": linkID}).
		RunWith(exec).ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *ChatLinkRepository) GetChatsByLinkID(ctx context.Context, linkID int64) ([]model.Chat, error) {
	exec := repository.GetExecutor(ctx, r.db)

	rows, err := r.psql.Select("c.ID, c.user_ID").From("chats c").
		Join("chat_link cl on cl.chat_ID = c.ID").
		Where(sq.Eq{"cl.link_ID": linkID}).RunWith(exec).QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing query builder chats by link rows", "error", closeErr)
		}
	}()
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

func (r *ChatLinkRepository) GetChatIDsByLinkID(ctx context.Context, linkID int64) ([]int64, error) {
	exec := repository.GetExecutor(ctx, r.db)

	rows, err := r.psql.Select("chat_ID").From("chat_link").Where(sq.Eq{"link_ID": linkID}).
		RunWith(exec).QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	chatIDs := make([]int64, 0)
	for rows.Next() {
		var chatID int64
		err = rows.Scan(&chatID)
		if err != nil {
			return nil, err
		}
		chatIDs = append(chatIDs, chatID)
	}
	if err = rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return chatIDs, nil
}
