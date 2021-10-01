//+build wireinject

package wire

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"

	"github.com/cloudnativedaysjp/reviewapp-operator/infrastructure/git"
	git_iface "github.com/cloudnativedaysjp/reviewapp-operator/infrastructure/git/iface"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/apprepo"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/infrarepo"
)

func NewGitRemoteRepoAppService(l logr.Logger) (*apprepo.GitRemoteRepoAppService, error) {
	wire.Build(
		git.NewGitApiPullRequestDriver,
		wire.Bind(new(git_iface.GitApiPullRequestIFace), new(*git.GitApiPullRequestDriver)),
		apprepo.NewGitRemoteRepoAppService,
	)
	return nil, nil
}

func NewGitRemoteRepoInfraService(l logr.Logger) (*infrarepo.GitRemoteRepoInfraService, error) {
	wire.Build(
		git.NewGitCommandDriver,
		wire.Bind(new(git_iface.GitCommandIFace), new(*git.GitCommandDriver)),
		infrarepo.NewGitRemoteRepoInfraService,
	)
	return nil, nil
}