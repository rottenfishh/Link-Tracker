//nolint:wrapcheck // repository methods mostly proxy storage errors from the DB layer
package rawsql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
)

type TagRepository struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) SaveTag(ctx context.Context, tag *model.Tag) (*model.Tag, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `INSERT INTO tags(name) 
            VALUES ($1) ON CONFLICT (name) DO UPDATE
            SET name = EXCLUDED.name
            RETURNING ID, name`
	row := exec.QueryRowContext(ctx, query, tag.Name)
	err := row.Scan(&tag.ID, &tag.Name)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (r *TagRepository) GetTags(ctx context.Context) ([]model.Tag, error) {
	query := `SELECT * FROM tags`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	tags := make([]model.Tag, 0)
	for rows.Next() {
		var tag model.Tag
		err = rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, err
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

	query := `DELETE FROM tags WHERE ID = $1 RETURNING ID, name   `
	row := exec.QueryRowContext(ctx, query, tagID)
	var tag model.Tag
	err := row.Scan(&tag.ID, &tag.Name)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) UpdateTag(ctx context.Context, tagID int64, tag *model.Tag) (*model.Tag, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `UPDATE tags SET name = $1 WHERE ID = $2
            RETURNING ID, name;`

	row := exec.QueryRowContext(ctx, query, tag.Name, tagID)

	var newTag model.Tag
	err := row.Scan(&newTag.ID, &newTag.Name)
	if err != nil {
		return nil, err
	}

	return &newTag, nil
}

func (r *TagRepository) DeleteTagByName(ctx context.Context, tagName string) (*model.Tag, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `DELETE FROM TAGS WHERE name = $1 RETURNING ID, name;`
	row := exec.QueryRowContext(ctx, query, tagName)
	var newTag model.Tag
	err := row.Scan(&newTag.ID, &newTag.Name)
	if err != nil {
		return nil, err
	}

	return &newTag, nil
}
