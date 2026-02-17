package app

import (
	"database/sql"
	"fmt"
	"github.com/dontpanicw/ImageProcessor/config"
	"github.com/dontpanicw/ImageProcessor/internal/adapter/broker"
	"github.com/dontpanicw/ImageProcessor/internal/adapter/repository/minio"
	"github.com/dontpanicw/ImageProcessor/internal/adapter/repository/postgres"
	"github.com/dontpanicw/ImageProcessor/internal/input/http"
	"github.com/dontpanicw/ImageProcessor/internal/usecases"
	"github.com/dontpanicw/ImageProcessor/pkg/migrations"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func Start(cfg *config.Config) error {
	// Retry подключения к PostgreSQL
	var db *sql.DB
	var err error

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
		return fmt.Errorf("failed to connect to database after 10 attempts: %w", err)
	}
	defer db.Close()
	log.Print("Connected to PostgreSQL")

	if err := migrations.Migrate(db); err != nil {
		fmt.Println(err)
		return err
	}
	log.Print("Migrations applied successfully")

	imageRepo := postgres.NewImageRepository(cfg)
	minioRepo := minio.NewMinioClient(cfg)

	// Инициализируем MinIO
	if err := minioRepo.InitMinio(); err != nil {
		return fmt.Errorf("failed to initialize MinIO: %w", err)
	}
	log.Print("MinIO initialized successfully")

	// Инициализируем Kafka producer
	kafkaProducer := broker.NewProducer(cfg)
	log.Print("Kafka producer initialized")

	// Даем Kafka время на инициализацию
	log.Print("Waiting for Kafka to be ready...")
	time.Sleep(5 * time.Second)

	imageUsecase := usecases.NewImageUsecases(imageRepo, minioRepo, kafkaProducer)

	srv := http.NewServer(cfg.HTTPPort, imageUsecase)

	return srv.Start()
}
