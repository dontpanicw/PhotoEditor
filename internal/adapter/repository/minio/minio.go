package minio

import (
	"context"
	"errors"
	"fmt"
	"github.com/dontpanicw/ImageProcessor/config"
	"github.com/dontpanicw/ImageProcessor/internal/port"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"time"
)

type ImageMinioStorage struct {
	mc     *minio.Client // Клиент Minio
	config *config.Config
}

// NewMinioClient создает новый экземпляр Minio Client
func NewMinioClient(cfg *config.Config) port.ObjectStorage {
	return &ImageMinioStorage{
		config: cfg,
	} // Возвращает новый экземпляр minioClient с указанным именем бакета
}

// InitMinio подключается к Minio и создает бакет, если не существует
// Бакет - это контейнер для хранения объектов в Minio. Он представляет собой пространство имен, в котором можно хранить и организовывать файлы и папки.
func (i *ImageMinioStorage) InitMinio() error {
	ctx := context.Background()

	var client *minio.Client
	var err error

	// пробуем 5 раз с паузой 3 секунды
	for k := 0; k < 5; k++ {
		client, err = minio.New(i.config.MinioEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(i.config.MinioRootUser, i.config.MinioRootPassword, ""),
			Secure: i.config.MinioUseSSL,
		})
		if err == nil {
			// сохранили клиент
			i.mc = client

			// проверяем бакет
			exists, err := i.mc.BucketExists(ctx, i.config.BucketName)
			if err == nil {
				if !exists {
					if err := i.mc.MakeBucket(ctx, i.config.BucketName, minio.MakeBucketOptions{}); err != nil {
						log.Printf("Ошибка при создании бакета: %v", err)
						return err
					}
					log.Printf("Бакет %s успешно создан", i.config.BucketName)
				} else {
					log.Printf("Бакет %s уже существует", i.config.BucketName)
				}
				return nil // всё успешно
			}
		}

		log.Printf("MinIO ещё не готов (попытка %d/10): %v", k+1, err)
		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("не удалось подключиться к MinIO после 10 попыток: %w", err)
}

func (i *ImageMinioStorage) PutObject(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) error {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	info, err := i.mc.PutObject(
		ctx,
		i.config.BucketName,
		objectKey,
		r,
		size,
		opts,
	)
	if err != nil {
		return fmt.Errorf("ошибка при загрузке объекта %s: %w", objectKey, err)
	}
	// Логируем успешную загрузку
	log.Printf("Объект %s успешно загружен в MinIO, размер: %d байт", objectKey, info.Size)
	return nil
}

func (i *ImageMinioStorage) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	// Получение предварительно подписанного URL для доступа к объекту Minio.
	object, err := i.mc.GetObject(
		ctx,
		i.config.BucketName,
		objectKey,
		minio.GetObjectOptions{},
	)
	if err != nil {
		log.Printf("ошибка при получении файла")
		return nil, fmt.Errorf("ошибка при получении объекта %s: %v", objectKey, err)
	}
	log.Println("файл получен из minio")
	return object, nil
}

func (i *ImageMinioStorage) RemoveObject(ctx context.Context, objectKey string) error {
	if objectKey == "" {
		return errors.New("object key cannot be empty")
	}

	err := i.mc.RemoveObject(ctx, i.config.BucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		// Проверяем, существует ли объект
		var minioErr minio.ErrorResponse
		if errors.As(err, &minioErr) && minioErr.Code == "NoSuchKey" {
			return fmt.Errorf("object %s not found: %w", objectKey, err)
		}
		return fmt.Errorf("failed to remove object %s: %w", objectKey, err)
	}

	log.Printf("Object %s successfully removed from MinIO", objectKey)
	return nil
}
