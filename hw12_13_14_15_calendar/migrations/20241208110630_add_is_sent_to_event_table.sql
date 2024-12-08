-- +goose Up
-- +goose StatementBegin
ALTER TABLE event
    ADD COLUMN is_sent BOOLEAN NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE event
    DROP COLUMN is_sent;
-- +goose StatementEnd
