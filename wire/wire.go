//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"k8s.io/utils/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/services"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/gitcommand"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/githubapi"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/kubernetes"
)

func NewGitHubAPIRepository(l logr.Logger) (*githubapi.GitHub, error) {
	wire.Build(
		githubapi.NewGitHub,
	)
	return nil, nil
}

func NewGitCommandRepository(l logr.Logger, e exec.Interface) (*gitcommand.Git, error) {
	wire.Build(
		gitcommand.NewGit,
	)
	return nil, nil
}

func NewKubernetesRepository(l logr.Logger, e client.Client) (*kubernetes.Client, error) {
	wire.Build(
		kubernetes.NewClient,
	)
	return nil, nil
}

func NewPullRequestService(l logr.Logger) (*services.PullRequestService, error) {
	wire.Build(
		githubapi.NewGitHub,
		wire.Bind(new(repositories.GitAPI), new(*githubapi.GitHub)),
		services.NewPullRequestService,
	)
	return nil, nil
}
