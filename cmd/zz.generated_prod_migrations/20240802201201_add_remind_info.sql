-- +goose Up
ALTER TABLE Tasks
ADD RemindedAt integer NOT NULL DEFAULT 0;
ALTER TABLE Tasks
ADD RemindAfter integer NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE Tasks
    DROP COLUMN RemindedAt;
ALTER TABLE Tasks
    DROP COLUMN RemindAfter;
