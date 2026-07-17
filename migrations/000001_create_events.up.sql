CREATE TABLE IF NOT EXISTS events(
    id SERIAL PRIMARY KEY,
    event_id INT UNIQUE NOT NULL,
    type TEXT NOT NULL,
    payload TEXT NOT NULL,
    status VARCHAR CHECK (status IN ('pending', 'processing', 'done', 'failed')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
)