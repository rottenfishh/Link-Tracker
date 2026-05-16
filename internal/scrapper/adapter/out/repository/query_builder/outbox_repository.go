package querybuilder

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
)

type OutboxRepository struct {
	db   *sql.DB
	psql sq.StatementBuilderType
}

func NewOutboxRepository(db *sql.DB, psql sq.StatementBuilderType) *OutboxRepository {
	return &OutboxRepository{db: db, psql: psql}
}

func (r *OutboxRepository) Save(ctx context.Context, outbox *model.Outbox) (*model.Outbox, error) {
	exec := repository.GetExecutor(ctx, r.db)

	var savedOutbox model.Outbox
	err := r.psql.Insert("outbox").Columns("id", "topic", "message_type", "payload", "created_at").
		Values(outbox.ID, outbox.Topic, outbox.MessageType, outbox.Payload, outbox.CreatedAt).
		Suffix("RETURNING id, topic, message_type, payload, created_at").
		RunWith(exec).QueryRowContext(ctx).Scan(&savedOutbox.ID, &savedOutbox.Topic, &savedOutbox.MessageType,
		&savedOutbox.Payload, &savedOutbox.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("saving outbox: %w", err)
	}

	return &savedOutbox, nil
}

func (r *OutboxRepository) GetPendingOutbox(ctx context.Context) ([]model.Outbox, error) {
	exec := repository.GetExecutor(ctx, r.db)

	rows, err := r.psql.Select("id, topic, message_type, payload, created_at").From("outbox").
		Where(sq.Eq{"processed_at": nil}).Suffix("FOR UPDATE SKIP LOCKED").RunWith(exec).QueryContext(ctx)

	if err != nil {
		return nil, fmt.Errorf("GetPendingOutbox: %w", err)
	}
	defer rows.Close() //nolint:errcheck //its ok bro

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

	var updOutbox model.Outbox

	err := r.psql.Update("outbox").Where(sq.Eq{"id": id}).Set("processed_at", processedAt).
		Suffix("RETURNING id, topic, message_type, payload, created_at, processed_at").
		RunWith(exec).QueryRowContext(ctx).
		Scan(&updOutbox.ID, &updOutbox.Topic, &updOutbox.MessageType,
			&updOutbox.Payload, &updOutbox.CreatedAt, &updOutbox.ProcessedAt)
	if err != nil {
		return nil, fmt.Errorf("update processed outbox error: %w", err)
	}

	return &updOutbox, nil
}

func (r *OutboxRepository) UpdateError(ctx context.Context, outboxID string, varError string) (*model.Outbox, error) {
	exec := repository.GetExecutor(ctx, r.db)

	var updatedOutbox model.Outbox
	err := r.psql.Update("outbox").Where(sq.Eq{"id": outboxID}).Set("error", varError).
		Suffix("RETURNING id, topic, message_type, payload, created_at").
		RunWith(exec).Scan(&updatedOutbox.ID, &updatedOutbox.Topic,
		&updatedOutbox.MessageType, &updatedOutbox.Payload, &updatedOutbox.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("update outbox error: %w", err)
	}

	return &updatedOutbox, nil
}

func (r *OutboxRepository) Delete(ctx context.Context, outboxID string) error {
	exec := repository.GetExecutor(ctx, r.db)

	_, err := r.psql.Delete("outbox").Where(sq.Eq{"id": outboxID}).RunWith(exec).ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("delete outbox error: %w", err)
	}

	return nil
}
