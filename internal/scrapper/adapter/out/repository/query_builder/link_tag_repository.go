//nolint:wrapcheck // repository methods mostly proxy storage errors and keep generated-like method grouping
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

type LinkTagRepository struct {
	db   *sql.DB
	psql sq.StatementBuilderType
}

func NewLinkTagRepository(db *sql.DB, psql sq.StatementBuilderType) *LinkTagRepository {
	return &LinkTagRepository{db: db, psql: psql}
}

func (r *LinkTagRepository) GetTagsByLinkID(ctx context.Context, linkID int64) ([]model.Tag, error) {
	rows, err := r.psql.Select("t.id, t.name").From("tags t").
		Join("link_tag cl on cl.tag_id = t.id").Where(sq.Eq{"cl.link_id": linkID}).
		RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck //its ok

	tags := make([]model.Tag, 0)
	for rows.Next() {
		var tag model.Tag
		if err = rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return tags, nil
}

func (r *LinkTagRepository) GetLinksByTagID(ctx context.Context, tagID int64) ([]model.Link, error) {
	rows, err := r.psql.Select("l.ID, l.link, l.domain, l.last_updated").From("links l").
		Join("link_tag cl on cl.link_ID = l.ID").Where(sq.Eq{"cl.tag_ID": tagID}).RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing query builder links by tag rows", "error", closeErr)
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

	_, err := r.psql.Insert("link_tag").Columns("link_ID", "tag_ID").Values(linkID, tagID).
		Suffix("ON CONFLICT DO NOTHING").RunWith(exec).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (r *LinkTagRepository) Delete(ctx context.Context, linkID int64, tagID int64) error {
	exec := repository.GetExecutor(ctx, r.db)

	_, err := r.psql.Delete("link_tag").Where(sq.Eq{"link_ID": linkID, "tag_ID": tagID}).
		RunWith(exec).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (r *LinkTagRepository) DeleteTagsByLinkID(ctx context.Context, linkID int64) error {
	exec := repository.GetExecutor(ctx, r.db)
	err := exec.QueryRowContext(ctx, `DELETE FROM link_tag WHERE link_ID = $1`, linkID).Err()
	if err != nil {
		return err
	}
	return nil
}
