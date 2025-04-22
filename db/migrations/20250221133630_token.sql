-- +goose Up
-- +goose StatementBegin
CREATE
    EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE table tokens (
    id serial PRIMARY KEY ,
    token TEXT NOT NULL unique,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMP DEFAULT current_timestamp + INTERVAL '7 day',
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists tokens;
-- +goose StatementEnd
