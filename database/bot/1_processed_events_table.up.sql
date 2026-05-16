CREATE TABLE IF NOT EXISTS processed_events(
    event_id uuid PRIMARY KEY,
    processed_at TIMESTAMP NOT NULL DEFAULT now()
)