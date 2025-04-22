-- +goose Up
-- +goose StatementBegin
CREATE
    EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS subscriptions (
    sub_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (sub_id, target_user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS subscriptions;
-- +goose StatementEnd
