-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS shortener AUTHORIZATION shortener;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
