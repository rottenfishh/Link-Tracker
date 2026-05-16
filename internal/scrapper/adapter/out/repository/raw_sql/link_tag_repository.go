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

type LinkTagRepository struct {
	db *sql.DB
}

func (r *LinkTagRepository) DeleteTagsByLinkID(ctx context.Context, linkID int64) error {
	exec := repository.GetExecutor(ctx, r.db)

	query := `DELETE FROM link_tag WHERE link_ID = $1`
	_, err := exec.ExecContext(ctx, query, linkID)
	if err != nil {
		return err
	}

	return nil
}

func NewLinkTagRepository(pool *sql.DB) *LinkTagRepository {
	return &LinkTagRepository{db: pool}
}

func (r *LinkTagRepository) GetLinksByTagID(ctx context.Context, tagID int64) ([]model.Link, error) {
	query := `SELECT l.ID, l.link, l.domain, l.last_updated
            FROM links l
            JOIN link_tag cl ON cl.link_ID = l.ID
            WHERE cl.tag_ID = $1`
	rows, err := r.db.QueryContext(ctx, query, tagID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing sql links by tag rows", "error", closeErr)
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

	if err = rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return links, nil
}

func (r *LinkTagRepository) Save(ctx context.Context, linkID int64, tagID int64) error {
	exec := repository.GetExecutor(ctx, r.db)

	query := `INSERT INTO link_tag(link_ID, tag_ID) VALUES($1, $2) ON CONFLICT DO NOTHING`
	_, err := exec.ExecContext(ctx, query, linkID, tagID)
	if err != nil {
		return err
	}
	return nil
}

func (r *LinkTagRepository) Delete(ctx context.Context, linkID int64, tagID int64) error {
	exec := repository.GetExecutor(ctx, r.db)

	query := `DELETE FROM link_tag WHERE link_ID = $1 AND tag_ID = $2`
	_, err := exec.ExecContext(ctx, query, linkID, tagID)
	if err != nil {
		return err
	}
	return nil
}
