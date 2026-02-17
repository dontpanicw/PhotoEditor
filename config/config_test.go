package config

import (
	"os"
	"testing"
)

func TestNewConfig_Defaults(t *testing.T) {
	// Очищаем переменные окружения
	os.Clearenv()

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.HTTPPort != DefaultHTTPPort {
		t.Errorf("Expected default HTTP port %s, got %s", DefaultHTTPPort, cfg.HTTPPort)
	}

	if cfg.MinioEndpoint != DefaultMinioEndpoint {
		t.Errorf("Expected default Minio endpoint %s, got %s", DefaultMinioEndpoint, cfg.MinioEndpoint)
	}

	if cfg.MinioUseSSL != false {
		t.Error("Expected MinioUseSSL to be false by default")
	}

	if cfg.KafkaTaskTopic != "image-tasks" {
		t.Errorf("Expected default Kafka topic 'image-tasks', got %s", cfg.KafkaTaskTopic)
	}
}

func TestNewConfig_CustomValues(t *testing.T) {
	os.Clearenv()
	os.Setenv("HTTP_PORT", "9090")
	os.Setenv("MINIO_ENDPOINT", "minio:9000")
	os.Setenv("MASTER_DSN", "postgres://test")
	os.Setenv("BUCKET_NAME", "test-bucket")
	os.Setenv("MINIO_ROOT_USER", "testuser")
	os.Setenv("MINIO_ROOT_PASSWORD", "testpass")
	os.Setenv("KAFKA_TASK_TOPIC", "custom-topic")
	os.Setenv("KAFKA_BROKERS", "kafka:9092")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.HTTPPort != ":9090" {
		t.Errorf("Expected HTTP port :9090, got %s", cfg.HTTPPort)
	}

	if cfg.MinioEndpoint != "minio:9000" {
		t.Errorf("Expected Minio endpoint minio:9000, got %s", cfg.MinioEndpoint)
	}

	if cfg.MasterDSN != "postgres://test" {
		t.Errorf("Expected MasterDSN postgres://test, got %s", cfg.MasterDSN)
	}

	if cfg.BucketName != "test-bucket" {
		t.Errorf("Expected bucket name test-bucket, got %s", cfg.BucketName)
	}

	if cfg.KafkaTaskTopic != "custom-topic" {
		t.Errorf("Expected Kafka topic custom-topic, got %s", cfg.KafkaTaskTopic)
	}
}

func TestNewConfig_PortFormatting(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"8080", ":8080"},
		{":8080", ":8080"},
		{"", DefaultHTTPPort},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			os.Clearenv()
			if tt.input != "" {
				os.Setenv("HTTP_PORT", tt.input)
			}

			cfg, err := NewConfig()
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if cfg.HTTPPort != tt.expected {
				t.Errorf("Expected port %s, got %s", tt.expected, cfg.HTTPPort)
			}
		})
	}
}
