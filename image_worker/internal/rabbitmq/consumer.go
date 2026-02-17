package rabbitmq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dontpanicw/ImageProcessor/config"
	workerPort "github.com/dontpanicw/ImageProcessor/image_worker/internal/port"
	"github.com/dontpanicw/ImageProcessor/image_worker/internal/processor"
	"github.com/dontpanicw/ImageProcessor/internal/domain"
	"github.com/dontpanicw/ImageProcessor/internal/port"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	minio  port.ObjectStorage
	repo   port.RepositoryDB
}

func NewConsumer(cfg *config.Config, minio port.ObjectStorage, repo port.RepositoryDB) workerPort.Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.KafkaBrokers,
		Topic:          cfg.KafkaTaskTopic,
		GroupID:        "image-workers",
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	return &Consumer{
		reader: reader,
		minio:  minio,
		repo:   repo,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	log.Println("Starting to consume messages from Kafka...")

	// Запускаем несколько воркеров
	const workerCount = 5
	errCh := make(chan error, workerCount)

	for i := 0; i < workerCount; i++ {
		go c.startWorker(ctx, i, errCh)
	}

	// Ждем первую ошибку или завершение контекста
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (c *Consumer) startWorker(ctx context.Context, id int, errCh chan<- error) {
	log.Printf("Worker %d started and waiting for messages", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopped", id)
			return
		default:
			// Читаем сообщение с таймаутом
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					log.Printf("Worker %d: context canceled", id)
					return
				}
				log.Printf("Worker %d: error fetching message: %v", id, err)
				time.Sleep(time.Second)
				continue
			}

			log.Printf("Worker %d received message: topic=%s, partition=%d, offset=%d, key=%s",
				id, msg.Topic, msg.Partition, msg.Offset, string(msg.Key))

			// Обрабатываем сообщение
			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("Worker %d failed to process message: %v", id, err)
				// Не коммитим сообщение при ошибке
				continue
			}

			// Коммитим сообщение после успешной обработки
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Worker %d: failed to commit message: %v", id, err)
			} else {
				log.Printf("Worker %d successfully processed and committed message", id)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	// Парсим сообщение из Kafka
	var task domain.TaskMessage
	if err := json.Unmarshal(msg.Value, &task); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Printf("Processing image %s with actions %v", task.ImageID, task.Actions)

	// 1. Получаем метаданные из БД
	image, err := c.repo.GetObjectByID(ctx, task.ImageID)
	if err != nil {
		return fmt.Errorf("failed to get image from DB: %w", err)
	}

	// 2. Загружаем оригинал из MinIO (получаем io.ReadCloser)
	originalFile, err := c.minio.GetObject(ctx, image.RawImageObjectKey)
	if err != nil {
		return fmt.Errorf("failed to get object from MinIO: %w", err)
	}
	defer func() {
		err = originalFile.Close()
		if err != nil {
			log.Printf("Failed to close original file: %v", err)
		}
	}()

	// 3. Читаем весь файл в []byte
	imageData, err := io.ReadAll(originalFile)
	if err != nil {
		return fmt.Errorf("failed to read image data: %w", err)
	}

	// 4. Определяем Content-Type (можно сохранять в БД или определять по магии)
	contentType := http.DetectContentType(imageData)

	// 5. Последовательно применяем все действия к []byte
	currentData := imageData
	for _, action := range task.Actions {
		currentData, err = c.applyAction(action, currentData)
		if err != nil {
			return fmt.Errorf("failed to apply action %s: %w", action, err)
		}
	}

	// 6. Генерируем ключ для обработанного файла
	processedObjectKey := fmt.Sprintf("processed/%s/%s_%d.jpg",
		task.ImageID,
		uuid.New().String(),
		time.Now().Unix(),
	)

	// 7. Сохраняем обработанный файл в MinIO
	reader := bytes.NewReader(currentData)

	err = c.minio.PutObject(
		ctx,
		processedObjectKey,
		reader,
		int64(len(currentData)),
		contentType,
	)
	if err != nil {
		return fmt.Errorf("failed to save processed image to MinIO: %w", err)
	}

	// 8. Обновляем статус в БД
	err = c.repo.UpdateProcessedImage(ctx, task.ImageID, processedObjectKey)
	if err != nil {
		// Cleanup: удаляем из MinIO, если БД не обновилась
		if cleanupErr := c.minio.RemoveObject(ctx, processedObjectKey); cleanupErr != nil {
			log.Printf("CRITICAL: Failed to cleanup MinIO after DB error: %v", cleanupErr)
		}
		return fmt.Errorf("failed to update DB: %w", err)
	}

	log.Printf("Successfully processed image %s, size: %d bytes, saved as %s",
		task.ImageID, len(currentData), processedObjectKey)

	return nil
}

// applyAction применяет одно действие к изображению (работает с []byte)
func (c *Consumer) applyAction(action string, imageData []byte) ([]byte, error) {
	switch action {
	case domain.ResizeAction:
		return processor.ResizeImage(imageData, 1600, 900)

	case domain.WatermarkAction:
		return processor.AddTextWatermark(imageData, "WildBerries")

	case domain.MiniatureGenerateAction:
		return processor.GenerateSmartThumbnail(imageData, 1600, 900)

	case domain.GrayscaleAction:
		return processor.ApplyGrayscale(imageData)

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (c *Consumer) Close() error {
	if c.reader != nil {
		log.Println("Closing Kafka reader...")
		return c.reader.Close()
	}
	return nil
}
