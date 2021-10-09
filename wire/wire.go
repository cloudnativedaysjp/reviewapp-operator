//+build wireinject

package wire

import (
	"github.com/cloudnativedaysjp/reviewapp-operator/services"
	"github.com/cloudnativedaysjp/reviewapp-operator/wrapper"
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"k8s.io/utils/exec"
)

func NewGitRemoteRepoAppService(l logr.Logger) (*services.GitRemoteRepoAppService, error) {
	wire.Build(
		wrapper.NewGitHubDriver,
		wire.Bind(new(wrapper.GitHubIFace), new(*wrapper.GitHub)),
		services.NewGitRemoteRepoAppService,
	)
	return nil, nil
}

func NewGitRemoteRepoInfraService(l logr.Logger, e exec.Interface) (*services.GitRemoteRepoInfraService, error) {
	wire.Build(
		wrapper.NewGitCommandDriver,
		wire.Bind(new(gitcommand.GitCommandIFace), new(*wrapper.Git)),
		services.NewGitRemoteRepoInfraService,
	)
	return nil, nil
}
