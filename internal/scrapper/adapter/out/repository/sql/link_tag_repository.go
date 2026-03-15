package sql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type LinkTagRepository struct {
	db *pgxpool.Pool
}

func NewLinkTagRepository(pool *pgxpool.Pool) *LinkTagRepository {
	return &LinkTagRepository{db: pool}
}

func (r *LinkTagRepository) GetLinksByTagID(ctx context.Context, tagId int64) ([]model.Link, error) {
	sql := `SELECT l.id, l.link, l.domain, l.last_updated
            FROM links l
            JOIN link_tag cl ON cl.link_id = l.id
            WHERE cl.tag_id = $1`
	rows, err := r.db.Query(ctx, sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]model.Link, 0)
	for rows.Next() {
		var link model.Link
		if err = rows.Scan(&link.Id, &link.Link, &link.Domain, &link.LastUpdated); err != nil {
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

func (r *LinkTagRepository) Save(ctx context.Context, linkId int64, tagId int64) error {
	sql := `INSERT INTO link_tag(link_id, tag_id) VALUES($1, $2)`
	_, err := r.db.Exec(ctx, sql, linkId, tagId)
	if err != nil {
		return err
	}
	return nil
}

func (r *LinkTagRepository) Delete(ctx context.Context, linkId int64, tagId int64) error {
	sql := `DELETE FROM link_tag WHERE link_id = $1 AND tag_id = $2`
	_, err := r.db.Exec(ctx, sql, linkId, tagId)
	if err != nil {
		return err
	}
	return nil
}
