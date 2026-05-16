//nolint:wrapcheck // repository methods mostly proxy storage errors from the DB layer
package rawsql

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
)

type LinkRepository struct {
	db *sql.DB
}

func NewLinkRepository(db *sql.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) AddLink(ctx context.Context, link model.Link) (*model.Link, error) {
	exec := repository.GetExecutor(ctx, r.db)

	timeNow := time.Now().UTC().Add(-24 * time.Hour)

	slog.Debug("processing link", "timeNow", timeNow.String(), "link", link.Link)
	query := `INSERT INTO links(link, domain, last_updated, title, formatted_link) 
            VALUES ($1, $2, $3, $4, $5) ON CONFLICT (link) DO UPDATE
            SET link = EXCLUDED.link
            RETURNING ID, link, domain, last_updated, formatted_link, title;`
	row := exec.QueryRowContext(ctx, query, link.Link, link.Domain, timeNow, link.Title, link.FormattedLink)

	var addedLink model.Link

	err := row.Scan(&addedLink.ID, &addedLink.Link, &addedLink.Domain, &addedLink.LastUpdated,
		&addedLink.FormattedLink, &addedLink.Title)
	if err != nil {
		return nil, err
	}
	return &addedLink, nil
}

func (r *LinkRepository) DeleteLink(ctx context.Context, linkID int64) (*model.Link, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `DELETE FROM links WHERE ID = $1 RETURNING *`
	rows := exec.QueryRowContext(ctx, query, linkID)

	var link model.Link

	err := rows.Scan(&link.ID, &link.Link, &link.Domain, &link.LastUpdated, &link.FormattedLink, &link.Title)
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

	query := `DELETE FROM links WHERE link= $1 RETURNING *`
	rows := exec.QueryRowContext(ctx, query, linkName)

	var link model.Link
	err := rows.Scan(&link.ID, &link.Link, &link.Domain, &link.LastUpdated, &link.FormattedLink, &link.Title)
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

	query := `UPDATE links SET link = $1, domain = $2, last_updated = $3, title = $5, formatted_link = $6 WHERE ID = $4
            RETURNING ID, link, domain, last_updated, title, formatted_link;`

	row := exec.QueryRowContext(ctx, query, link.Link, link.Domain, link.LastUpdated, linkID, link.Title, link.FormattedLink)

	var linkNew model.Link
	err := row.Scan(&linkNew.ID, &linkNew.Link, &linkNew.Domain, &linkNew.LastUpdated, &linkNew.Title, &linkNew.FormattedLink)
	if err != nil {
		return nil, err
	}

	return &linkNew, nil
}

func (r *LinkRepository) GetLinks(ctx context.Context, offset, limit int64) ([]model.Link, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `SELECT * FROM links OFFSET $1 LIMIT $2;`
	rows, err := exec.QueryContext(ctx, query, offset, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing sql links rows", "error", closeErr)
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
	exec := repository.GetExecutor(ctx, r.db)

	query := `SELECT ID, link, domain, last_updated, formatted_link, title
		 FROM links
		 WHERE link = $1`

	row := exec.QueryRowContext(ctx, query, linkName)

	var link model.Link
	err := row.Scan(&link.ID, &link.Link, &link.Domain, &link.LastUpdated, &link.FormattedLink, &link.Title)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *LinkRepository) GetLinksOlderThan(ctx context.Context, time time.Time, limit, offset int) ([]model.Link, error) {

	query := `SELECT * FROM links WHERE last_updated < $1 ORDER BY last_updated LIMIT $2 OFFSET $3;`
	rows, err := r.db.QueryContext(ctx, query, time, limit, offset)
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
		return nil, rowsErr
	}
	return links, nil
}

func (r *LinkTagRepository) GetTagsByLinkID(ctx context.Context, linkID int64) ([]model.Tag, error) {
	query := `SELECT t.id, t.name
              FROM tags t
              JOIN link_tag cl on cl.tag_id = t.id
              WHERE cl.link_id = $1`
	rows, err := r.db.QueryContext(ctx, query, linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck //its ok

	var tags []model.Tag
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
