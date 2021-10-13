//+build wireinject

package wire

import (
	"github.com/cloudnativedaysjp/reviewapp-operator/services"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"k8s.io/utils/exec"
)

func NewGitRemoteRepoAppService(l logr.Logger) (*services.GitRemoteRepoAppService, error) {
	wire.Build(
		gateways.NewGitHubDriver,
		wire.Bind(new(gateways.GitHubIFace), new(*gateways.GitHub)),
		services.NewGitRemoteRepoAppService,
	)
	return nil, nil
}

func NewGitRemoteRepoInfraService(l logr.Logger, e exec.Interface) (*services.GitRemoteRepoInfraService, error) {
	wire.Build(
		gateways.NewGitCommandDriver,
		wire.Bind(new(gitcommand.GitCommandIFace), new(*gateways.Git)),
		services.NewGitRemoteRepoInfraService,
	)
	return nil, nil
}
