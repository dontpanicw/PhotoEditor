package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dontpanicw/ImageProcessor/config"
	"github.com/dontpanicw/ImageProcessor/image_worker/internal/rabbitmq"
	"github.com/dontpanicw/ImageProcessor/internal/adapter/repository/minio"
	"github.com/dontpanicw/ImageProcessor/internal/adapter/repository/postgres"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting Image Worker...")

	// Загружаем конфигурацию
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к PostgreSQL с retry
	var db *sql.DB
	
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", cfg.MasterDSN)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Waiting for PostgreSQL... (attempt %d/10): %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	
	if err != nil {
		log.Fatalf("Failed to connect to database after 10 attempts: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	// Инициализируем репозитории
	imageRepo := postgres.NewImageRepository(cfg)
	minioClient := minio.NewMinioClient(cfg)

	// Инициализируем MinIO bucket
	if err := minioClient.InitMinio(); err != nil {
		log.Fatalf("Failed to initialize MinIO bucket: %v", err)
	}
	log.Println("MinIO initialized")

	// Даем Kafka время на инициализацию
	log.Println("Waiting for Kafka to be ready...")
	time.Sleep(10 * time.Second)

	// Создаем Kafka consumer
	consumer := rabbitmq.NewConsumer(cfg, minioClient, imageRepo)
	log.Println("Kafka consumer created")

	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для обработки сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Запускаем consumer в отдельной горутине
	errChan := make(chan error, 1)
	go func() {
		log.Println("Starting Kafka consumer...")
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Consumer start error: %v", err)
			errChan <- err
		}
	}()

	// Ждем сигнал завершения или ошибку
	select {
	case <-sigChan:
		log.Println("Received shutdown signal")
	case err := <-errChan:
		log.Printf("Consumer error: %v", err)
	}

	// Graceful shutdown
	log.Println("Shutting down worker...")
	cancel()

	// Даем время на завершение обработки текущих сообщений
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Error closing consumer: %v", err)
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("Worker stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout exceeded")
	}
}
