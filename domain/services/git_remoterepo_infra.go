package services

import (
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
)

type GitRemoteRepoInfraService struct {
	prRepo     repositories.PullRequestAppIFace
	secretRepo repositories.GitRepoCredentialIFace
}

func NewGitRemoteRepoInfraService(pr repositories.PullRequestInfraIFace, secret repositories.GitRepoCredentialIFace) *GitRemoteRepoInfraService {
	return &GitRemoteRepoInfraService{pr, secret}
}

// TODO
