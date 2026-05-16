package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type ProcessedEventsRepository struct {
	db *sql.DB
}

func NewProcessedEventsRepository(db *sql.DB) *ProcessedEventsRepository {
	return &ProcessedEventsRepository{db: db}
}

// Register если мы получили true, nil - новая запись. если false, nil - запись уже была. иначе ошибка
func (r *ProcessedEventsRepository) Register(ctx context.Context, id string) (bool, error) {
	query := `INSERT INTO processed_events(event_id) 
              VALUES ($1) 
              ON CONFLICT DO NOTHING`

	row, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("error registering processed event: %w", err)
	}

	result, err := row.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("error registering processed event: %w", err)
	}

	if result == 0 {
		return false, nil
	}

	return true, nil
}

func (r *ProcessedEventsRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM processed_events WHERE event_id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return fmt.Errorf("error deleting processed event: %w", err)
}
