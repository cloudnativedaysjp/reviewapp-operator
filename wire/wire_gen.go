// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package wire

import (
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/gitcommand"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/githubapi"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/kubernetes"
	"github.com/go-logr/logr"
	"k8s.io/utils/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Injectors from wire.go:

func NewGitHubAPIRepository(l logr.Logger) (*githubapi.GitHub, error) {
	gitHub := githubapi.NewGitHub(l)
	return gitHub, nil
}

func NewGitCommandRepository(l logr.Logger, e exec.Interface) (*gitcommand.Git, error) {
	git, err := gitcommand.NewGit(l, e)
	if err != nil {
		return nil, err
	}
	return git, nil
}

func NewKubernetesRepository(l logr.Logger, e client.Client) (*kubernetes.Client, error) {
	kubernetesClient := kubernetes.NewClient(l, e)
	return kubernetesClient, nil
}
