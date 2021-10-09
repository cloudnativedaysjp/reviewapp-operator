//+build wireinject

package wire

import (
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	"github.com/cloudnativedaysjp/reviewapp-operator/services"
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"k8s.io/utils/exec"
)

func NewGitRemoteRepoAppService(l logr.Logger) (*services.GitRemoteRepoAppService, error) {
	wire.Build(
		gateways.NewGitHubApiDriver,
		wire.Bind(new(gateways.GitHubApiIFace), new(*gateways.GitHubApiDriver)),
		services.NewGitRemoteRepoAppService,
	)
	return nil, nil
}

func NewGitRemoteRepoInfraService(l logr.Logger, e exec.Interface) (*services.GitRemoteRepoInfraService, error) {
	wire.Build(
		gateways.NewGitCommandDriver,
		wire.Bind(new(gitcommand.GitCommandIFace), new(*gateways.GitCommandDriver)),
		services.NewGitRemoteRepoInfraService,
	)
	return nil, nil
}
