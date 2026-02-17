package port

import (
	"context"
)

type Producer interface {
	SendMessage(ctx context.Context, imageId string, actions []string) error
}
