// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package wire

import (
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	"github.com/cloudnativedaysjp/reviewapp-operator/services"
	"github.com/go-logr/logr"
	"k8s.io/utils/exec"
)

// Injectors from wire.go:

func NewGitRemoteRepoAppService(l logr.Logger) (*services.GitRemoteRepoAppService, error) {
	gitApiDriver := gateways.NewGitApiDriver(l)
	gitRemoteRepoAppService := services.NewGitRemoteRepoAppService(gitApiDriver)
	return gitRemoteRepoAppService, nil
}

func NewGitRemoteRepoInfraService(l logr.Logger, e exec.Interface) (*services.GitRemoteRepoInfraService, error) {
	gitCommandDriver, err := gateways.NewGitCommandDriver(l, e)
	if err != nil {
		return nil, err
	}
	gitRemoteRepoInfraService := services.NewGitRemoteRepoInfraService(gitCommandDriver)
	return gitRemoteRepoInfraService, nil
}
