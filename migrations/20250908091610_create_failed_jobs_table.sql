-- +goose Up
CREATE TABLE failed_jobs (
    id SERIAL PRIMARY KEY,
    upload_id VARCHAR(255) NOT NULL,
    job_type VARCHAR(50) NOT NULL,
    last_error VARCHAR(255) NOT NULL,
    payload BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    job_status VARCHAR(20) DEFAULT 'failed'
);

-- +goose Down
DROP TABLE IF EXISTS failed_jobs;
