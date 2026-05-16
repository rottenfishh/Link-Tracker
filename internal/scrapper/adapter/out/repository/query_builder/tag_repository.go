//nolint:wrapcheck // repository methods mostly proxy storage errors from the DB layer
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

type TagRepository struct {
	db   *sql.DB
	psql sq.StatementBuilderType
}

// TODO:
// db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
// if err != nil {
// panic("Unable to connect to database")
// }
//
// psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func NewTagRepository(db *sql.DB, psql sq.StatementBuilderType) *TagRepository {
	return &TagRepository{db: db, psql: psql}
}

func (r *TagRepository) SaveTag(ctx context.Context, tag *model.Tag) (*model.Tag, error) {
	exec := repository.GetExecutor(ctx, r.db)

	var savedTag model.Tag
	err := r.psql.Insert("tags").
		Columns("name").
		Values(tag.Name).
		Suffix("ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING ID, name").
		RunWith(exec).QueryRowContext(ctx).
		Scan(&savedTag.ID, &savedTag.Name)
	if err != nil {
		return nil, err
	}
	return &savedTag, nil
}

func (r *TagRepository) GetTags(ctx context.Context) ([]model.Tag, error) {
	rows, err := r.psql.Select("ID", "name").From("tags").RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Error("closing query builder tags rows", "error", closeErr)
		}
	}()

	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if scanErr := rows.Scan(&tag.ID, &tag.Name); scanErr != nil {
			return nil, scanErr
		}
		tags = append(tags, tag)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		if errors.Is(rowsErr, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, rowsErr
	}

	return tags, nil
}

func (r *TagRepository) DeleteTag(ctx context.Context, tagID int64) (*model.Tag, error) {
	exec := repository.GetExecutor(ctx, r.db)

	var tag model.Tag
	err := r.psql.Delete("tags").
		Where(sq.Eq{"ID": tagID}).
		Suffix("RETURNING ID, name").
		PlaceholderFormat(sq.Dollar).
		RunWith(exec).QueryRowContext(ctx).
		Scan(&tag.ID, &tag.Name)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) UpdateTag(ctx context.Context, tagID int64, tag *model.Tag) (*model.Tag, error) {
	var updatedTag model.Tag
	err := r.psql.Update("tags").
		Set("name", tag.Name).
		Where(sq.Eq{"ID": tagID}).
		Suffix("RETURNING ID, name").
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&updatedTag.ID, &updatedTag.Name)
	if err != nil {
		return nil, err
	}
	return &updatedTag, nil
}

func (r *TagRepository) DeleteTagByName(ctx context.Context, tagName string) (*model.Tag, error) {
	var tag model.Tag
	err := r.psql.Delete("tags").
		Where(sq.Eq{"name": tagName}).
		Suffix("RETURNING ID, name").
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&tag.ID, &tag.Name)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}
