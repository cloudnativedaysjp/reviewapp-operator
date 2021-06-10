package githubapi

import (
	"github.com/go-logr/logr"

	"github.com/cloudnativedaysjp/reviewapp-operator/services/repositories"
)

type GitApiFactoryImpl struct{}

func (kfi GitApiFactoryImpl) NewRepository(l logr.Logger, username, token string) (repositories.GitApiRepository, error) {
	return NewGitHubApiInfra(l, username, token)
}
