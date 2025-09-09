package migrations

import (
	"database/sql"
	"fmt"
)

// +goose Up
// +goose StatementBegin
func Up(tx *sql.Tx) error {
	// Image tablosu:
	createImageTable := `
	CREATE TABLE images (
		id UUID PRIMARY KEY,
		original_name VARCHAR(255) NOT NULL,
		file_type VARCHAR(50) NOT NULL,
		file_path VARCHAR(500) NOT NULL,
		status VARCHAR(20) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);
	`
	if _, err := tx.Exec(createImageTable); err != nil {
		return fmt.Errorf("could not create images table: %w", err)
	}

	// MediaVariant tablosu:
	createMediaVariantTable := `
	CREATE TABLE media_variants (
		variant_id UUID PRIMARY KEY,
		media_id UUID NOT NULL,
		variant_name VARCHAR(255) NOT NULL,
		width INTEGER NOT NULL,
		height INTEGER NOT NULL,
		file_path VARCHAR(500) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	`
	if _, err := tx.Exec(createMediaVariantTable); err != nil {
		return fmt.Errorf("could not create media_variants table: %w", err)
	}

	// MediaSize tablosu:
	createMediaSizeTable := `
	CREATE TABLE media_sizes (
		variant_type VARCHAR(255) PRIMARY KEY,
		width INTEGER NOT NULL,
		height INTEGER NOT NULL
	);
	`
	if _, err := tx.Exec(createMediaSizeTable); err != nil {
		return fmt.Errorf("could not create media_sizes table: %w", err)
	}

	// Video tablosu:
	createVideoTable := `
	CREATE TABLE videos (
		video_id UUID PRIMARY KEY,
		original_name VARCHAR(255) NOT NULL,
		file_type VARCHAR(50) NOT NULL,
		file_path VARCHAR(500) NOT NULL,
		status VARCHAR(50) NOT NULL,
		height BIGINT NOT NULL,
		width BIGINT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	`
	if _, err := tx.Exec(createVideoTable); err != nil {
		return fmt.Errorf("could not create videos table: %w", err)
	}

	createFailedJobsTable := `
	CREATE TABLE failed_jobs (
		id SERIAL PRIMARY KEY,
		upload_id VARCHAR(255) NOT NULL,
		job_type VARCHAR(50) NOT NULL,
		last_error VARCHAR(255) NOT NULL,
		payload BYTEA NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		job_status VARCHAR(20) DEFAULT 'failed'
	);
	`
	if _, err := tx.Exec(createFailedJobsTable); err != nil {
		return fmt.Errorf("could not create failed_jobs table: %w", err)
	}

	return nil
}

// +goose StatementEnd

// +goose Down
// +goose StatementBegin
func Down(tx *sql.Tx) error {
	// Tabloları silme işlemini ters sırada yap.
	dropTables := []string{"videos", "media_sizes", "media_variants", "images"}
	for _, table := range dropTables {
		if _, err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)); err != nil {
			return fmt.Errorf("could not drop table %s: %w", table, err)
		}
	}
	return nil
}

// +goose StatementEnd
