package port

import (
	"context"
	"github.com/dontpanicw/ImageProcessor/internal/domain"
	"io"
)

type RepositoryDB interface {
	SaveObject(ctx context.Context, image domain.Image) error
	GetObjectByID(ctx context.Context, id string) (*domain.Image, error)
	DeleteObjectByID(ctx context.Context, id string) error
	UpdateProcessedImage(ctx context.Context, id string, processedObjectKey string) error
}

type ObjectStorage interface {
	InitMinio() error
	PutObject(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error)
	RemoveObject(ctx context.Context, objectKey string) error
}

//встроенные HTTP-методы:
//– POST /upload — загрузка изображения на обработку;
//– GET /image/{id} — получение обработанного изображения;
//– DELETE /image/{id} — удаление изображения.
