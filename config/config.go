package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	HTTPPort          string
	MasterDSN         string
	SlaveDSNs         []string
	MinioEndpoint     string // Адрес конечной точки Minio
	BucketName        string // Название конкретного бакета в Minio
	MinioRootUser     string // Имя пользователя для доступа к Minio
	MinioRootPassword string // Пароль для доступа к Minio
	MinioUseSSL       bool
	KafkaTaskTopic    string
	KafkaBrokers      []string
}

const (
	DefaultHTTPPort      = ":8080"
	DefaultMinioEndpoint = ":9000"
)

func NewConfig() (*Config, error) {
	cfg := Config{
		MinioEndpoint: DefaultMinioEndpoint,
		MinioUseSSL:   false,
	}

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		cfg.HTTPPort = DefaultHTTPPort
	} else {
		// Ensure port starts with ':' if not already present
		if len(httpPort) > 0 && httpPort[0] != ':' {
			cfg.HTTPPort = ":" + httpPort
		} else {
			cfg.HTTPPort = httpPort
		}
	}
	
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint != "" {
		cfg.MinioEndpoint = minioEndpoint
	}
	
	masterDSN := os.Getenv("MASTER_DSN")
	if masterDSN != "" {
		cfg.MasterDSN = masterDSN
	}
	slaveDSNs := make([]string, 0)
	slaveDSN := os.Getenv("SLAVE_DSN")
	slaveDSNs = append(slaveDSNs, slaveDSN)

	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName != "" {
		cfg.BucketName = bucketName
	}
	minioRootUser := os.Getenv("MINIO_ROOT_USER")
	if minioRootUser != "" {
		cfg.MinioRootUser = minioRootUser
	}
	minioRootPassword := os.Getenv("MINIO_ROOT_PASSWORD")
	if minioRootPassword != "" {
		cfg.MinioRootPassword = minioRootPassword
	}
	kafkaTaskTopic := os.Getenv("KAFKA_TASK_TOPIC")
	if kafkaTaskTopic != "" {
		cfg.KafkaTaskTopic = kafkaTaskTopic
	} else {
		cfg.KafkaTaskTopic = "image-tasks"
	}
	
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers != "" {
		cfg.KafkaBrokers = []string{kafkaBrokers}
	} else {
		cfg.KafkaBrokers = []string{"localhost:9092"}
	}

	return &cfg, nil
}
