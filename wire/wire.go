//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"k8s.io/utils/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/gitcommand"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/githubapi"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/kubernetes"
)

func NewGitHubApi(l logr.Logger) (*githubapi.GitHub, error) {
	wire.Build(
		githubapi.NewGitHub,
	)
	return nil, nil
}

func NewGitLocalRepo(l logr.Logger, e exec.Interface) (*gitcommand.GitLocalRepo, error) {
	wire.Build(
		gitcommand.NewGitLocalRepo,
	)
	return nil, nil
}

func NewKubernetes(l logr.Logger, e client.Client) (*kubernetes.Client, error) {
	wire.Build(
		kubernetes.NewClient,
	)
	return nil, nil
}
