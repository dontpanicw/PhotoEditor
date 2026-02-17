package usecases

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/dontpanicw/ImageProcessor/internal/domain"
)

// Mock implementations
type mockRepositoryDB struct {
	saveObjectFunc       func(ctx context.Context, image domain.Image) error
	getObjectByIDFunc    func(ctx context.Context, id string) (*domain.Image, error)
	deleteObjectByIDFunc func(ctx context.Context, id string) error
	updateProcessedFunc  func(ctx context.Context, id string, key string) error
}

func (m *mockRepositoryDB) SaveObject(ctx context.Context, image domain.Image) error {
	if m.saveObjectFunc != nil {
		return m.saveObjectFunc(ctx, image)
	}
	return nil
}

func (m *mockRepositoryDB) GetObjectByID(ctx context.Context, id string) (*domain.Image, error) {
	if m.getObjectByIDFunc != nil {
		return m.getObjectByIDFunc(ctx, id)
	}
	return &domain.Image{Id: id, Status: domain.ImageStatusDone}, nil
}

func (m *mockRepositoryDB) DeleteObjectByID(ctx context.Context, id string) error {
	if m.deleteObjectByIDFunc != nil {
		return m.deleteObjectByIDFunc(ctx, id)
	}
	return nil
}

func (m *mockRepositoryDB) UpdateProcessedImage(ctx context.Context, id string, key string) error {
	if m.updateProcessedFunc != nil {
		return m.updateProcessedFunc(ctx, id, key)
	}
	return nil
}

type mockObjectStorage struct {
	initMinioFunc    func() error
	putObjectFunc    func(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	getObjectFunc    func(ctx context.Context, key string) (io.ReadCloser, error)
	removeObjectFunc func(ctx context.Context, key string) error
}

func (m *mockObjectStorage) InitMinio() error {
	if m.initMinioFunc != nil {
		return m.initMinioFunc()
	}
	return nil
}

func (m *mockObjectStorage) PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, key, r, size, contentType)
	}
	return nil
}

func (m *mockObjectStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, key)
	}
	return io.NopCloser(strings.NewReader("test")), nil
}

func (m *mockObjectStorage) RemoveObject(ctx context.Context, key string) error {
	if m.removeObjectFunc != nil {
		return m.removeObjectFunc(ctx, key)
	}
	return nil
}

type mockProducer struct {
	sendMessageFunc func(ctx context.Context, imageId string, actions []string) error
}

func (m *mockProducer) SendMessage(ctx context.Context, imageId string, actions []string) error {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, imageId, actions)
	}
	return nil
}

func TestCreateObject_Success(t *testing.T) {
	repo := &mockRepositoryDB{}
	storage := &mockObjectStorage{}
	producer := &mockProducer{}

	usecase := NewImageUsecases(repo, storage, producer)

	image := domain.Image{
		FileName: "test.jpg",
		FileSize: 1024,
		Actions:  []string{domain.ResizeAction},
	}

	reader := strings.NewReader("test image data")
	ctx := context.Background()

	id, err := usecase.CreateObject(ctx, image, reader, 1024, "image/jpeg")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if id == "" {
		t.Fatal("Expected non-empty ID")
	}
}

func TestCreateObject_ValidationError(t *testing.T) {
	repo := &mockRepositoryDB{}
	storage := &mockObjectStorage{}
	producer := &mockProducer{}

	usecase := NewImageUsecases(repo, storage, producer)

	image := domain.Image{
		FileName: "", // Invalid: empty filename
		FileSize: 1024,
	}

	reader := strings.NewReader("test")
	ctx := context.Background()

	_, err := usecase.CreateObject(ctx, image, reader, 1024, "image/jpeg")

	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}
}

func TestCreateObject_MinioError(t *testing.T) {
	repo := &mockRepositoryDB{}
	storage := &mockObjectStorage{
		putObjectFunc: func(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
			return errors.New("minio error")
		},
	}
	producer := &mockProducer{}

	usecase := NewImageUsecases(repo, storage, producer)

	image := domain.Image{
		FileName: "test.jpg",
		FileSize: 1024,
		Actions:  []string{domain.ResizeAction},
	}

	reader := strings.NewReader("test")
	ctx := context.Background()

	_, err := usecase.CreateObject(ctx, image, reader, 1024, "image/jpeg")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCreateObject_DBError_WithCleanup(t *testing.T) {
	cleanupCalled := false

	repo := &mockRepositoryDB{
		saveObjectFunc: func(ctx context.Context, image domain.Image) error {
			return errors.New("db error")
		},
	}
	storage := &mockObjectStorage{
		removeObjectFunc: func(ctx context.Context, key string) error {
			cleanupCalled = true
			return nil
		},
	}
	producer := &mockProducer{}

	usecase := NewImageUsecases(repo, storage, producer)

	image := domain.Image{
		FileName: "test.jpg",
		FileSize: 1024,
		Actions:  []string{domain.ResizeAction},
	}

	reader := strings.NewReader("test")
	ctx := context.Background()

	_, err := usecase.CreateObject(ctx, image, reader, 1024, "image/jpeg")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !cleanupCalled {
		t.Fatal("Expected cleanup to be called")
	}
}

func TestGetObjectByID_Success(t *testing.T) {
	repo := &mockRepositoryDB{
		getObjectByIDFunc: func(ctx context.Context, id string) (*domain.Image, error) {
			return &domain.Image{
				Id:                      id,
				Status:                  domain.ImageStatusDone,
				ProcessedImageObjectKey: "processed/test.jpg",
			}, nil
		},
	}
	storage := &mockObjectStorage{}
	producer := &mockProducer{}

	usecase := NewImageUsecases(repo, storage, producer)
	ctx := context.Background()

	reader, err := usecase.GetObjectByID(ctx, "test-id")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if reader == nil {
		t.Fatal("Expected reader, got nil")
	}
	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(reader)
}

func TestGetObjectByID_PendingStatus(t *testing.T) {
	repo := &mockRepositoryDB{
		getObjectByIDFunc: func(ctx context.Context, id string) (*domain.Image, error) {
			return &domain.Image{
				Id:     id,
				Status: domain.ImageStatusPending,
			}, nil
		},
	}
	storage := &mockObjectStorage{}
	producer := &mockProducer{}

	usecase := NewImageUsecases(repo, storage, producer)
	ctx := context.Background()

	_, err := usecase.GetObjectByID(ctx, "test-id")

	if err == nil {
		t.Fatal("Expected error for pending status, got nil")
	}
}

func TestGetObjectByID_FailedStatus(t *testing.T) {
	repo := &mockRepositoryDB{
		getObjectByIDFunc: func(ctx context.Context, id string) (*domain.Image, error) {
			return &domain.Image{
				Id:     id,
				Status: domain.ImageStatusFailed,
			}, nil
		},
	}
	storage := &mockObjectStorage{}
	producer := &mockProducer{}

	usecase := NewImageUsecases(repo, storage, producer)
	ctx := context.Background()

	_, err := usecase.GetObjectByID(ctx, "test-id")

	if err == nil {
		t.Fatal("Expected error for failed status, got nil")
	}
}

func TestRemoveObject_Success(t *testing.T) {
	repo := &mockRepositoryDB{
		getObjectByIDFunc: func(ctx context.Context, id string) (*domain.Image, error) {
			return &domain.Image{
				Id:                      id,
				RawImageObjectKey:       "raw/test.jpg",
				ProcessedImageObjectKey: "processed/test.jpg",
			}, nil
		},
	}

	minioCallCount := 0
	storage := &mockObjectStorage{
		removeObjectFunc: func(ctx context.Context, key string) error {
			minioCallCount++
			return nil
		},
	}
	producer := &mockProducer{}

	usecase := NewImageUsecases(repo, storage, producer)
	ctx := context.Background()

	err := usecase.RemoveObject(ctx, "test-id")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if minioCallCount != 2 {
		t.Fatalf("Expected 2 minio calls, got %d", minioCallCount)
	}
}

func TestValidateImage(t *testing.T) {
	tests := []struct {
		name    string
		image   domain.Image
		wantErr bool
	}{
		{
			name: "valid image",
			image: domain.Image{
				FileName: "test.jpg",
				FileSize: 1024,
			},
			wantErr: false,
		},
		{
			name: "empty filename",
			image: domain.Image{
				FileName: "",
				FileSize: 1024,
			},
			wantErr: true,
		},
		{
			name: "zero file size",
			image: domain.Image{
				FileName: "test.jpg",
				FileSize: 0,
			},
			wantErr: true,
		},
		{
			name: "negative file size",
			image: domain.Image{
				FileName: "test.jpg",
				FileSize: -100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImage(&tt.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
