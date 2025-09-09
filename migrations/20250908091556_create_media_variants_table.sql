-- +goose Up
CREATE TABLE media_variants (
    variant_id UUID PRIMARY KEY,
    media_id UUID NOT NULL,
    variant_name VARCHAR(255) NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS media_variants;
