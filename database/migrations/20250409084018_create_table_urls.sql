-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
id uuid PRIMARY KEY,
short_key varchar(32) NOT NULL UNIQUE,
original_url varchar(256) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd
