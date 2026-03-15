package sql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type LinkRepository struct {
	db *sql.DB
}

func NewLinkRepository(db *sql.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) AddLink(ctx context.Context, link model.Link) (*model.Link, error) {
	sql := `INSERT INTO links(link, domain) 
            VALUES ($1, $2) ON CONFLICT (link) DO UPDATE
            SET link = EXCLUDED.link
            RETURNING id, link, domain, last_updated;`
	row := r.db.QueryRowContext(ctx, sql, link.Link, link.Domain)

	var addedLink model.Link

	err := row.Scan(&addedLink.Id, &addedLink.Link, &addedLink.Domain, &addedLink.LastUpdated)
	if err != nil {
		return nil, err
	}
	return &addedLink, nil
}

func (r *LinkRepository) DeleteLink(ctx context.Context, linkID int64) (*model.Link, error) {
	sql := `DELETE FROM links WHERE id = $1 RETURNING *`
	rows := r.db.QueryRowContext(ctx, sql, linkID)

	var link model.Link

	err := rows.Scan(&link.Id, &link.Link, &link.Domain, &link.LastUpdated)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *LinkRepository) DeleteLinkByName(ctx context.Context, linkName string) (*model.Link, error) {
	sql := `DELETE FROM links WHERE link= $1 RETURNING *`
	rows := r.db.QueryRowContext(ctx, sql, linkName)

	var link model.Link
	err := rows.Scan(&link.Id, &link.Link, &link.Domain, &link.LastUpdated)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (r *LinkRepository) UpdateLink(ctx context.Context, linkID int64, link *model.Link) (*model.Link, error) {
	sql := `UPDATE links SET link = $1, domain = $2, last_updated = now() WHERE id = $3
            RETURNING id, link, domain, last_updated;`

	row := r.db.QueryRowContext(ctx, sql, link.Link, link.Domain, linkID)

	var linkNew model.Link
	err := row.Scan(&linkNew.Id, &linkNew.Link, &linkNew.Domain, &linkNew.LastUpdated)
	if err != nil {
		return nil, err
	}

	return &linkNew, nil
}

func (r *LinkRepository) GetLinks(ctx context.Context) ([]model.Link, error) {
	sql := `SELECT * FROM links`
	rows, err := r.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var links []model.Link
	for rows.Next() {
		var link model.Link
		err = rows.Scan(&link.Id, &link.Link, &link.Domain, &link.LastUpdated)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return links, nil
}

func (r *LinkRepository) GetLinkByName(ctx context.Context, linkName string) (*model.Link, error) {
	sql := `SELECT id, link, domain, last_updated
		 FROM links
		 WHERE link = $1`

	row := r.db.QueryRowContext(ctx, sql, linkName)

	var link model.Link
	err := row.Scan(&link.Id, &link.Link, &link.Domain, &link.LastUpdated)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}
