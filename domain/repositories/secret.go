package repositories

import (
	"context"
)

type SecretIFace interface {
	GetSecretValue(ctx context.Context, namespace string, name string, key string) (string, error)
}
