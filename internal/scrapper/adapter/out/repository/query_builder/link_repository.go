package query_builder

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type LinkRepository struct {
	db   *sql.DB
	psql sq.StatementBuilderType
}

func NewLinkRepository(db *sql.DB, psql sq.StatementBuilderType) *LinkRepository {
	return &LinkRepository{db: db, psql: psql}
}

func (r *LinkRepository) AddLink(ctx context.Context, link model.Link) (*model.Link, error) {
	var addedLink model.Link
	err := r.psql.Insert("links").Columns("link", "domain").
		Values(link.Link, link.Domain).Suffix(
		"ON CONFLICT(link) DO UPDATE SET link = EXCLUDED.link RETURNING id, link, domain, last_updated;").
		RunWith(r.db).QueryRowContext(ctx).Scan(&addedLink.Id, &addedLink.Link, &addedLink.Domain, &addedLink.LastUpdated)
	if err != nil {
		return nil, err
	}

	return &addedLink, nil
}

func (r *LinkRepository) DeleteLink(ctx context.Context, linkID int64) (*model.Link, error) {
	var link model.Link
	err := r.psql.Delete("links").Where(sq.Eq{"id": linkID}).
		RunWith(r.db).QueryRowContext(ctx).Scan(&link.Id, &link.Link, &link.Domain, &link.LastUpdated)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (r *LinkRepository) DeleteLinkByName(ctx context.Context, linkName string) (*model.Link, error) {
	var link model.Link
	err := r.psql.Delete("links").Where(sq.Eq{"link": linkName}).
		Suffix("RETURNING *").QueryRowContext(ctx).Scan(&link.Id, &link.Link, &link.Domain, &link.LastUpdated)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (r *LinkRepository) UpdateLink(ctx context.Context, linkID int64, link *model.Link) (*model.Link, error) {
	var linkNew model.Link
	err := r.psql.Update("links").Set("link", link.Link).
		Set("domain", link.Domain).Set("last_updated", time.Now()).Where(sq.Eq{"id": linkID}).
		Suffix("RETURNING *").RunWith(r.db).QueryRowContext(ctx).
		Scan(&linkNew.Id, &linkNew.Link, &linkNew.Domain, &linkNew.LastUpdated)

	if err != nil {
		return nil, err
	}

	return &linkNew, nil
}

func (r *LinkRepository) GetLinks(ctx context.Context, offset, limit int64) ([]model.Link, error) {
	rows, err := r.psql.Select("*").From("links").Offset(uint64(offset)).Limit(uint64(limit)).RunWith(r.db).QueryContext(ctx)
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
	var link model.Link
	err := r.psql.Select("*").From("links").Where(sq.Eq{"link": linkName}).
		RunWith(r.db).QueryRowContext(ctx).Scan(&link.Id, &link.Link, &link.Domain, &link.LastUpdated)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (r *LinkRepository) GetLinksOlderThan(ctx context.Context, time time.Time, limit, offset int) ([]model.Link, error) {
	rows, err := r.psql.Select("*").From("links").Where(sq.Lt{"last_updated": time}).
		OrderBy("last_updated ASC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	links := make([]model.Link, 0)
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
