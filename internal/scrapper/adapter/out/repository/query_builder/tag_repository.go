package query_builder

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type TagRepository struct {
	db   *sql.DB
	psql sq.StatementBuilderType
}

//TODO:
//db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
//if err != nil {
//panic("Unable to connect to database")
//}
//
//psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func NewTagRepository(db *sql.DB, psql sq.StatementBuilderType) *TagRepository {
	return &TagRepository{db: db, psql: psql}
}

func (r *TagRepository) SaveTag(ctx context.Context, tag *model.Tag) (*model.Tag, error) {
	var savedTag model.Tag
	err := r.psql.Insert("tags").
		Columns("name").
		Values(tag.Name).
		Suffix("ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id, name").
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&savedTag.Id, &savedTag.Name)
	if err != nil {
		return nil, err
	}
	return &savedTag, nil
}

func (r *TagRepository) GetTags(ctx context.Context) ([]model.Tag, error) {
	rows, err := r.psql.Select("id", "name").From("tags").RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return tags, nil
}

func (r *TagRepository) DeleteTag(ctx context.Context, tagId int64) (*model.Tag, error) {
	var tag model.Tag
	err := r.psql.Delete("tags").
		Where(sq.Eq{"id": tagId}).
		Suffix("RETURNING id, name").
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&tag.Id, &tag.Name)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) UpdateTag(ctx context.Context, tagId int64, tag *model.Tag) (*model.Tag, error) {
	var updatedTag model.Tag
	err := r.psql.Update("tags").
		Set("name", tag.Name).
		Where(sq.Eq{"id": tagId}).
		Suffix("RETURNING id, name").
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&updatedTag.Id, &updatedTag.Name)
	if err != nil {
		return nil, err
	}
	return &updatedTag, nil
}

func (r *TagRepository) DeleteTagByName(ctx context.Context, tagName string) (*model.Tag, error) {
	var tag model.Tag
	err := r.psql.Delete("tags").
		Where(sq.Eq{"name": tagName}).
		Suffix("RETURNING id, name").
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&tag.Id, &tag.Name)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}
