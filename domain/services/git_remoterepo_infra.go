package services

import (
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
)

type GitRemoteRepoInfraService struct {
	prRepo     repositories.PullRequestInfraIFace
	secretRepo repositories.GitRepoSecretIFace
}

func NewGitRemoteRepoInfraService(pr repositories.PullRequestInfraIFace, secret repositories.GitRepoSecretIFace) *GitRemoteRepoInfraService {
	return &GitRemoteRepoInfraService{pr, secret}
}

// TODO
