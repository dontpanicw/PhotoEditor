package port

import (
	"context"
	"github.com/dontpanicw/ImageProcessor/internal/domain"
	"io"
)

type ImageUsecases interface {
	InitMinio() error
	CreateObject(ctx context.Context, image domain.Image, r io.Reader, size int64, contentType string) (string, error)
	GetObjectByID(ctx context.Context, id string) (io.ReadCloser, error)
	GetImageStatus(ctx context.Context, id string) (*domain.Image, error)
	RemoveObject(ctx context.Context, id string) error
}
