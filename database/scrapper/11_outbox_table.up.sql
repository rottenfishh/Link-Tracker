CREATE TABLE IF NOT EXISTS outbox(
    id uuid PRIMARY KEY ,
    topic VARCHAR(255) NOT NULL,
    message_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL ,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    processed_at TIMESTAMP,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_outbox_not_processed
    ON outbox(processed_at)
    WHERE processed_at IS NULL;