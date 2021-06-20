package repositories

import (
	"context"
)

type GitRepoSecret struct {
	Iface GitRepoSecretIFace
}

type GitRepoSecretIFace interface {
	GetSecretValue(ctx context.Context, namespace string, name string, key string) (string, error)
}

func NewGitRepoSecret(iface GitRepoSecretIFace) *GitRepoSecret {
	return &GitRepoSecret{iface}
}
