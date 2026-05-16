//nolint:wrapcheck,funcorder // repository methods mostly proxy storage errors and keep generated-like method grouping
package rawsql

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
)

type ChatLinkRepository struct {
	db *sql.DB
}

func (r *ChatLinkRepository) GetLinksByChatIDAndTag(ctx context.Context, chatID int64, tag string) ([]model.Link, error) {
	query := `SELECT l.ID, l.link, l.domain, l.last_updated
            FROM links l
            JOIN chat_link cl on cl.link_ID = l.ID
            JOIN link_tag lt on lt.link_ID = l.ID
            JOIN tags t on t.ID = lt.tag_ID
            WHERE cl.chat_ID = $1 AND t.name = $2`
	rows, err := r.db.QueryContext(ctx, query, chatID, tag)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing query chat links rows", "error", closeErr)
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
	return links, nil
}

func NewChatLinkRepository(db *sql.DB) *ChatLinkRepository {
	return &ChatLinkRepository{db: db}
}

func (r *ChatLinkRepository) GetLinksByChatID(ctx context.Context, chatID int64) ([]model.Link, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `SELECT l.ID, l.link, l.domain, l.last_updated
            FROM links l
            JOIN chat_link cl ON cl.link_ID = l.ID
            WHERE cl.chat_ID = $1`
	rows, err := exec.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing query links by chat rows", "error", closeErr)
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

	query := `INSERT INTO chat_link(chat_ID, link_ID) VALUES ($1, $2)`
	_, err := exec.ExecContext(ctx, query, chatID, linkID)
	if err != nil {
		return err
	}
	return nil
}

func (r *ChatLinkRepository) Unsubscribe(ctx context.Context, chatID int64, linkID int64) error {
	exec := repository.GetExecutor(ctx, r.db)

	query := "DELETE FROM chat_link WHERE chat_ID = $1 AND link_ID = $2"
	_, err := exec.ExecContext(ctx, query, chatID, linkID)
	if err != nil {
		return err
	}
	return nil
}

func (r *ChatLinkRepository) GetChatsByLinkID(ctx context.Context, linkID int64) ([]model.Chat, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `SELECT c.ID, c.user_ID
            FROM chats c
            JOIN chat_link cl ON cl.chat_ID = c.ID
            WHERE cl.link_ID = $1`
	rows, err := exec.QueryContext(ctx, query, linkID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing query chats by link rows", "error", closeErr)
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

	query := "SELECT chat_ID FROM chat_link WHERE link_ID = $1"
	rows, err := exec.QueryContext(ctx, query, linkID)
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
