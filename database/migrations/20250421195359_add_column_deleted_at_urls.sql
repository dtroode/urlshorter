-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls
ADD deleted_at TIMESTAMP WITH TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls
DROP COLUMN deleted_at;
-- +goose StatementEnd
