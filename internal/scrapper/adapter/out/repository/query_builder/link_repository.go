//nolint:wrapcheck // repository methods mostly proxy storage errors from the DB layer
package querybuilder

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
)

type LinkRepository struct {
	db   *sql.DB
	psql sq.StatementBuilderType
}

func NewLinkRepository(db *sql.DB, psql sq.StatementBuilderType) *LinkRepository {
	return &LinkRepository{db: db, psql: psql}
}

func (r *LinkRepository) AddLink(ctx context.Context, link model.Link) (*model.Link, error) {
	exec := repository.GetExecutor(ctx, r.db)

	timeNow := time.Now().UTC().Add(-24 * time.Hour)

	slog.Debug("processing link", "timeNow", timeNow.String(), "link", link.Link)
	var addedLink model.Link
	err := r.psql.Insert("links").Columns("link", "domain", "last_updated", "title", "formatted_link").
		Values(link.Link, link.Domain, timeNow, link.Title, link.FormattedLink).Suffix(
		"ON CONFLICT(link) DO UPDATE SET link = EXCLUDED.link RETURNING ID, link, domain, last_updated, title, formatted_link;").
		RunWith(exec).QueryRowContext(ctx).Scan(&addedLink.ID, &addedLink.Link, &addedLink.Domain,
		&addedLink.LastUpdated, &addedLink.FormattedLink, &addedLink.Title)
	if err != nil {
		return nil, err
	}

	return &addedLink, nil
}

func (r *LinkRepository) DeleteLink(ctx context.Context, linkID int64) (*model.Link, error) {
	exec := repository.GetExecutor(ctx, r.db)

	var link model.Link
	err := r.psql.Delete("links").Where(sq.Eq{"ID": linkID}).Suffix("RETURNING ID, link, domain, last_updated, formatted_link;").
		RunWith(exec).QueryRowContext(ctx).Scan(&link.ID, &link.Link, &link.Domain, &link.LastUpdated, &link.FormattedLink)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (r *LinkRepository) DeleteLinkByName(ctx context.Context, linkName string) (*model.Link, error) {
	exec := repository.GetExecutor(ctx, r.db)

	var link model.Link
	err := r.psql.Delete("links").Where(sq.Eq{"link": linkName}).
		Suffix("RETURNING *").RunWith(exec).QueryRowContext(ctx).
		Scan(&link.ID, &link.Link, &link.Domain, &link.LastUpdated, &link.FormattedLink, &link.Title)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (r *LinkRepository) UpdateLink(ctx context.Context, linkID int64, link *model.Link) (*model.Link, error) {
	exec := repository.GetExecutor(ctx, r.db)

	var linkNew model.Link
	err := r.psql.Update("links").Set("link", link.Link).
		Set("domain", link.Domain).Set("last_updated", link.LastUpdated).Set("formatted_link", link.FormattedLink).Where(sq.Eq{"ID": linkID}).
		Suffix("RETURNING *").RunWith(exec).QueryRowContext(ctx).
		Scan(&linkNew.ID, &linkNew.Link, &linkNew.Domain, &linkNew.LastUpdated,
			&linkNew.FormattedLink, &link.Title)

	if err != nil {
		return nil, err
	}

	return &linkNew, nil
}

func (r *LinkRepository) GetLinks(ctx context.Context, offset, limit int64) ([]model.Link, error) {
	rows, err := r.psql.Select("*").From("links").Offset(uint64(offset)).
		Limit(uint64(limit)).RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing query builder links rows", "error", closeErr)
		}
	}()
	var links []model.Link
	for rows.Next() {
		var link model.Link
		err = rows.Scan(&link.ID, &link.Link, &link.Domain, &link.LastUpdated, &link.FormattedLink, &link.Title)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		if errors.Is(rowsErr, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, rowsErr
	}
	return links, nil
}

func (r *LinkRepository) GetLinkByName(ctx context.Context, linkName string) (*model.Link, error) {
	var link model.Link
	err := r.psql.Select("*").From("links").Where(sq.Eq{"link": linkName}).
		RunWith(r.db).QueryRowContext(ctx).Scan(&link.ID, &link.Link, &link.Domain,
		&link.LastUpdated, &link.FormattedLink, &link.Title)

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
		err = rows.Scan(&link.ID, &link.Link, &link.Domain, &link.LastUpdated,
			&link.FormattedLink, &link.Title)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		if errors.Is(rowsErr, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, rowsErr
	}
	return links, nil
}
