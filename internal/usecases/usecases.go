package usecases

import (
	"context"
	"errors"
	"fmt"
	"github.com/dontpanicw/ImageProcessor/internal/domain"
	"github.com/dontpanicw/ImageProcessor/internal/port"
	"github.com/google/uuid"
	"io"
	"log"
)

var _ port.ImageUsecases = (*ImageUsecases)(nil)

type ImageUsecases struct {
	repo           port.RepositoryDB
	minio          port.ObjectStorage
	brokerProducer port.Producer
}

func NewImageUsecases(repo port.RepositoryDB, minio port.ObjectStorage, brokerProducer port.Producer) *ImageUsecases {
	return &ImageUsecases{
		repo:           repo,
		minio:          minio,
		brokerProducer: brokerProducer,
	}
}

func (i *ImageUsecases) InitMinio() error {
	return i.minio.InitMinio()
}

func (i *ImageUsecases) CreateObject(ctx context.Context, image domain.Image, r io.Reader, size int64, contentType string) (string, error) {
	//Валидация
	if err := validateImage(&image); err != nil {
		return "", fmt.Errorf("invalid image data: %w", err)
	}

	id := uuid.New().String()
	image.Id = id

	// Генерируем ключ для сырого изображения
	rawObjectKey := fmt.Sprintf("raw/%s/%s", image.Id, uuid.New().String())
	image.RawImageObjectKey = rawObjectKey
	image.Status = domain.ImageStatusPending

	log.Printf("Uploading image to MinIO: %s", rawObjectKey)
	err := i.minio.PutObject(ctx, rawObjectKey, r, size, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	log.Printf("Saving image metadata to DB: %s", image.Id)
	err = i.repo.SaveObject(ctx, image)
	if err != nil {
		if cleanupErr := i.minio.RemoveObject(ctx, rawObjectKey); cleanupErr != nil {
			log.Printf("CRITICAL: Failed to cleanup minio object %s after DB error: %v", rawObjectKey, cleanupErr)
		}
		return "", fmt.Errorf("failed to save to database: %w", err)
	}

	log.Printf("Sending task to Kafka for image: %s", image.Id)
	err = i.brokerProducer.SendMessage(ctx, image.Id, image.Actions)
	if err != nil {
		log.Printf("ERROR: Failed to send message to Kafka: %v", err)
		return "", fmt.Errorf("failed to send task to Kafka: %w", err)
	}
	
	log.Printf("Image %s successfully queued for processing", image.Id)
	return image.Id, nil
}

func (i *ImageUsecases) GetObjectByID(ctx context.Context, id string) (io.ReadCloser, error) {

	imageData, err := i.repo.GetObjectByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if imageData.Status == domain.ImageStatusPending {
		return nil, errors.New("image is pending")
	}
	if imageData.Status == domain.ImageStatusFailed {
		return nil, errors.New("image processing is failed")
	}

	image, err := i.minio.GetObject(ctx, imageData.ProcessedImageObjectKey)
	if err != nil {
		return nil, errors.New("error get object from minio")
	}

	return image, err
}

func (i *ImageUsecases) GetImageStatus(ctx context.Context, id string) (*domain.Image, error) {
	return i.repo.GetObjectByID(ctx, id)
}

func (i *ImageUsecases) RemoveObject(ctx context.Context, id string) error {
	imageData, err := i.repo.GetObjectByID(ctx, id)
	if err != nil {
		return err
	}
	err = i.repo.DeleteObjectByID(ctx, id)
	if err != nil {
		return err
	}
	err = i.minio.RemoveObject(ctx, imageData.ProcessedImageObjectKey)
	if err != nil {
		return err
	}
	err = i.minio.RemoveObject(ctx, imageData.RawImageObjectKey)
	if err != nil {
		return err
	}
	return nil
}

func validateImage(image *domain.Image) error {
	if image.FileName == "" {
		return errors.New("filename is required")
	}
	if image.FileSize <= 0 {
		return errors.New("file size must be positive")
	}
	return nil
}
