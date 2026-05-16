CREATE TABLE IF NOT EXISTS links(
    id SERIAL PRIMARY KEY,
    link VARCHAR(500) NOT NULL UNIQUE,
    domain VARCHAR(50) NOT NULL,
    last_updated timestamp default now()
)