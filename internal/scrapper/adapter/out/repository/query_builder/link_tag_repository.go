package query_builder

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type LinkTagRepository struct {
	db   *sql.DB
	psql sq.StatementBuilderType
}

func NewLinkTagRepository(db *sql.DB, psql sq.StatementBuilderType) *LinkTagRepository {
	return &LinkTagRepository{db: db, psql: psql}
}

func (r *LinkTagRepository) GetLinksByTagID(ctx context.Context, tagId int64) ([]model.Link, error) {
	rows, err := r.psql.Select("l.id, l.link, l.domain, l.last_updated").From("links l").
		Join("link_tag cl on cl.link_id = l.id").Where(sq.Eq{"cl.tag_id": tagId}).RunWith(r.db).QueryContext(ctx)
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
	_, err := r.psql.Insert("link_tag").Columns("link_id", "tag_id").Values(linkId, tagId).
		RunWith(r.db).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (r *LinkTagRepository) Delete(ctx context.Context, linkId int64, tagId int64) error {
	_, err := r.psql.Delete("link_tag").Where(sq.Eq{"link_id": linkId, "tag_id": tagId}).
		RunWith(r.db).Exec()
	if err != nil {
		return err
	}

	return nil
}
