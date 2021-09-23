// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package wire

import (
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/services"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/gitcommand"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/githubapi"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/kubernetes"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Injectors from wire.go:

func NewGitRemoteRepoAppService(l logr.Logger, c client.Client, username string) (*services.GitRemoteRepoAppService, error) {
	gitHubApiGateway := githubapi.NewGitHubApiGateway(l, username)
	kubernetesGateway, err := kubernetes.NewKubernetesGateway(c, l)
	if err != nil {
		return nil, err
	}
	gitRemoteRepoAppService := services.NewGitRemoteRepoAppService(gitHubApiGateway, kubernetesGateway, l)
	return gitRemoteRepoAppService, nil
}

func NewGitRemoteRepoInfraService(l logr.Logger, c client.Client, username string) (*services.GitRemoteRepoInfraService, error) {
	gitCommandGateway, err := gitcommand.NewGitCommandGateway(l, username)
	if err != nil {
		return nil, err
	}
	kubernetesGateway, err := kubernetes.NewKubernetesGateway(c, l)
	if err != nil {
		return nil, err
	}
	gitRemoteRepoInfraService := services.NewGitRemoteRepoInfraService(gitCommandGateway, kubernetesGateway, l)
	return gitRemoteRepoInfraService, nil
}

func NewKubernetesService(l logr.Logger, c client.Client) (*services.KubernetesService, error) {
	kubernetesGateway, err := kubernetes.NewKubernetesGateway(c, l)
	if err != nil {
		return nil, err
	}
	kubernetesService := services.NewKubernetes(kubernetesGateway, kubernetesGateway, kubernetesGateway, kubernetesGateway, l)
	return kubernetesService, nil
}
