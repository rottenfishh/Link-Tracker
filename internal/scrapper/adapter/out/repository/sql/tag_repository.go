package sql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type TagRepository struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) SaveTag(ctx context.Context, tag *model.Tag) (*model.Tag, error) {
	sql := `INSERT INTO tags(name) 
            VALUES ($1) ON CONFLICT (name) DO UPDATE
            SET name = EXCLUDED.name
            RETURNING id, name`
	row := r.db.QueryRowContext(ctx, sql, tag.Name)
	err := row.Scan(&tag.Id, &tag.Name)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (r *TagRepository) GetTags(ctx context.Context) ([]model.Tag, error) {
	sql := `SELECT * FROM tags`
	rows, err := r.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	tags := make([]model.Tag, 0)
	for rows.Next() {
		var tag model.Tag
		err = rows.Scan(&tag.Id, &tag.Name)
		if err != nil {
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
	sql := `DELETE FROM tags WHERE id = $1 RETURNING id, name   `
	row := r.db.QueryRowContext(ctx, sql, tagId)
	var tag model.Tag
	err := row.Scan(&tag.Id, &tag.Name)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) UpdateTag(ctx context.Context, tagId int64, tag *model.Tag) (*model.Tag, error) {
	sql := `UPDATE tags SET name = $1 WHERE id = $2
            RETURNING id, name;`

	row := r.db.QueryRowContext(ctx, sql, tag.Name, tagId)

	var newTag model.Tag
	err := row.Scan(&newTag.Id, &newTag.Name)
	if err != nil {
		return nil, err
	}

	return &newTag, nil
}

func (r *TagRepository) DeleteTagByName(ctx context.Context, tagName string) (*model.Tag, error) {
	sql := `DELETE FROM TAGS WHERE name = $1 RETURNING id, name;`
	row := r.db.QueryRowContext(ctx, sql, tagName)
	var newTag model.Tag
	err := row.Scan(&newTag.Id, &newTag.Name)
	if err != nil {
		return nil, err
	}

	return &newTag, nil
}
