-- +goose Up
CREATE TABLE media_sizes (
    variant_type VARCHAR(255) PRIMARY KEY,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS media_sizes;
