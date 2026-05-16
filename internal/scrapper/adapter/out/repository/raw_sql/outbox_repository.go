package rawsql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Save(ctx context.Context, outbox *model.Outbox) (*model.Outbox, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `INSERT INTO outbox(id, topic, message_type, payload, created_at)
              VALUES ($1, $2, $3, $4, $5)
              RETURNING id, topic, message_type, payload, created_at`

	var savedOutbox model.Outbox
	err := exec.QueryRowContext(ctx, query, outbox.ID, outbox.Topic,
		outbox.MessageType, outbox.Payload, outbox.CreatedAt).
		Scan(&savedOutbox.ID, &savedOutbox.Topic, &savedOutbox.MessageType,
			&savedOutbox.Payload, &savedOutbox.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("saving outbox: %w", err)
	}

	return &savedOutbox, nil
}

func (r *OutboxRepository) GetPendingOutbox(ctx context.Context) ([]model.Outbox, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `SELECT id, topic, message_type, payload, created_at
              FROM outbox
              WHERE processed_at IS NULL
              FOR UPDATE SKIP LOCKED`

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetPendingOutbox: %w", err)
	}
	defer rows.Close() //nolint:errcheck //idc

	var itemsOutbox []model.Outbox
	for rows.Next() {
		var item model.Outbox
		err = rows.Scan(&item.ID, &item.Topic, &item.MessageType, &item.Payload, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetPendingOutbox: %w", err)
		}
		itemsOutbox = append(itemsOutbox, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPendingOutbox: %w", err)
	}

	return itemsOutbox, nil
}

func (r *OutboxRepository) UpdateProcessedAt(ctx context.Context, id string, processedAt time.Time) (*model.Outbox, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `UPDATE outbox SET processed_at = $1 WHERE id = $2
              RETURNING id, topic, message_type, payload, created_at, processed_at`

	var updatedOutbox model.Outbox

	err := exec.QueryRowContext(ctx, query, processedAt, id).
		Scan(&updatedOutbox.ID, &updatedOutbox.Topic, &updatedOutbox.MessageType,
			&updatedOutbox.Payload, &updatedOutbox.CreatedAt, &updatedOutbox.ProcessedAt)
	if err != nil {
		return nil, fmt.Errorf("update processed outbox error: %w", err)
	}

	return &updatedOutbox, nil
}

func (r *OutboxRepository) UpdateError(ctx context.Context, outboxID string, errorVar string) (*model.Outbox, error) {
	exec := repository.GetExecutor(ctx, r.db)

	query := `UPDATE outbox SET error = $1 WHERE id = $2
              RETURNING id, topic, message_type, payload, created_at`

	var updatedOutbox model.Outbox
	err := exec.QueryRowContext(ctx, query, errorVar, outboxID).
		Scan(&updatedOutbox.ID, &updatedOutbox.Topic, &updatedOutbox.MessageType,
			&updatedOutbox.Payload, &updatedOutbox.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("update outbox errorVar: %w", err)
	}

	return &updatedOutbox, nil
}

func (r *OutboxRepository) Delete(ctx context.Context, outboxID string) error {
	exec := repository.GetExecutor(ctx, r.db)

	query := `DELETE FROM outbox WHERE id = $1 `
	_, err := exec.ExecContext(ctx, query, outboxID)
	if err != nil {
		return fmt.Errorf("delete outbox error: %w", err)
	}
	return nil
}
