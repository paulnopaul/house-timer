-- +goose Up
CREATE TABLE remind (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    CreatedAt INTEGER,
    DeletedAt INTEGER,

    ChatID INT NOT NULL,
    CurrentTask INT,

    RemindCount INT
);
-- +goose Down
DROP TABLE IF EXISTS remind;