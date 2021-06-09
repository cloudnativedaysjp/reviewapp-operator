package githubapi

import (
	"github.com/go-logr/logr"

	"github.com/cloudnativedaysjp/reviewapp-operator/services/repositories"
)

type GitApiFactoryImpl struct{}

func (kfi GitApiFactoryImpl) NewRepository(username, token string, l logr.Logger) (repositories.GitApiRepository, error) {
	return NewGitHubApiInfra(username, token, l)
}
