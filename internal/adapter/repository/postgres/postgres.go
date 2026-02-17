package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dontpanicw/ImageProcessor/config"
	"github.com/dontpanicw/ImageProcessor/internal/domain"
	"github.com/dontpanicw/ImageProcessor/internal/port"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"time"
)

type ImageRepository struct {
	PostgresDB *dbpg.DB
}

func NewImageRepository(cfg *config.Config) port.RepositoryDB {
	opts := &dbpg.Options{MaxOpenConns: 10, MaxIdleConns: 5}
	db, err := dbpg.New(cfg.MasterDSN, cfg.SlaveDSNs, opts)
	if err != nil {
		panic(err)
	}

	return &ImageRepository{
		PostgresDB: db,
	}
}

func (i *ImageRepository) SaveObject(ctx context.Context, image domain.Image) error {
	query := `
        INSERT INTO images (
            id, filename, file_size, raw_image_object_key, 
            processed_image_object_key, actions, status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (id) DO UPDATE SET
            filename = EXCLUDED.filename,
            file_size = EXCLUDED.file_size,
            raw_image_object_key = EXCLUDED.raw_image_object_key,
            processed_image_object_key = EXCLUDED.processed_image_object_key,
            actions = EXCLUDED.actions,
            status = EXCLUDED.status
    `

	// Сериализуем actions в JSON (теперь это просто массив строк)
	actionsJSON, err := json.Marshal(image.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	_, err = i.PostgresDB.ExecWithRetry(ctx, createRetryStrategy(), query,
		image.Id,
		image.FileName,
		image.FileSize,
		image.RawImageObjectKey,
		image.ProcessedImageObjectKey,
		actionsJSON,
		image.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}

func (i *ImageRepository) GetObjectByID(ctx context.Context, id string) (*domain.Image, error) {
	query := `
        SELECT 
            id, 
            filename, 
            file_size, 
            raw_image_object_key, 
            processed_image_object_key, 
            actions, 
            status 
        FROM images 
        WHERE id = $1
    `

	var image domain.Image
	var actionsJSON []byte

	err := i.PostgresDB.QueryRowContext(ctx, query, id).Scan(
		&image.Id,
		&image.FileName,
		&image.FileSize,
		&image.RawImageObjectKey,
		&image.ProcessedImageObjectKey,
		&actionsJSON,
		&image.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("image with id %s not found: %w", id, domain.ErrImageNotFound)
		}
		return nil, fmt.Errorf("failed to get image by id %s: %w", id, err)
	}

	// Десериализуем actions из JSONB в []string
	if err := json.Unmarshal(actionsJSON, &image.Actions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal actions for image %s: %w", id, err)
	}

	return &image, nil
}

func (i *ImageRepository) DeleteObjectByID(ctx context.Context, id string) error {
	query := `DELETE FROM images WHERE id = $1`

	result, err := i.PostgresDB.ExecWithRetry(ctx, createRetryStrategy(), query, id)
	if err != nil {
		return fmt.Errorf("failed to delete image %s: %w", id, err)
	}

	// Проверяем, была ли удалена запись
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for image %s: %w", id, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: id=%s", domain.ErrImageNotFound, id)
	}

	return nil
}

func (i *ImageRepository) UpdateProcessedImage(ctx context.Context, id string, processedObjectKey string) error {
	query := `UPDATE images 
              SET status = $1, 
                  processed_image_object_key = $2
              WHERE id = $3`

	_, err := i.PostgresDB.ExecContext(ctx, query, domain.ImageStatusDone, processedObjectKey, id)
	return err

}

func createRetryStrategy() retry.Strategy {
	return retry.Strategy{
		Attempts: 3,
		Delay:    5 * time.Second,
		Backoff:  2}
}
