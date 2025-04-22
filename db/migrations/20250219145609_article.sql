-- +goose Up
-- +goose StatementBegin
CREATE TABLE article
(
    slug VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    body TEXT NOT NULL,
    taglist TEXT[], -- Убедитесь, что тип TEXT[] поддерживается вашей СУБД
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    favoritesCount INT NOT NULL DEFAULT 0, -- Убрано SELECT, добавлено статическое значение
    author VARCHAR(255) NOT NULL,
    CONSTRAINT fk_author FOREIGN KEY (author) REFERENCES users(username)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS article;
-- +goose StatementEnd

