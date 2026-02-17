-- +goose Up
CREATE TABLE IF NOT EXISTS images (
                                      id VARCHAR(255) PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size > 0),
    raw_image_object_key VARCHAR(512) NOT NULL,
    processed_image_object_key VARCHAR(512),
    actions JSONB NOT NULL DEFAULT '[]',
    status VARCHAR(50) NOT NULL DEFAULT 'Pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

     -- Ограничения (constraints)
     CONSTRAINT valid_status CHECK (status IN ('Pending', 'Done', 'Failed'))
    );
