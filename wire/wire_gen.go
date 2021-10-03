// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package wire

import (
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/gitapi"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/gitcommand"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/apprepo"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/infrarepo"
	"github.com/go-logr/logr"
	"k8s.io/utils/exec"
)

// Injectors from wire.go:

func NewGitRemoteRepoAppService(l logr.Logger) (*apprepo.GitRemoteRepoAppService, error) {
	gitApiDriver := gitapi.NewGitApiDriver(l)
	gitRemoteRepoAppService := apprepo.NewGitRemoteRepoAppService(gitApiDriver, l)
	return gitRemoteRepoAppService, nil
}

func NewGitRemoteRepoInfraService(l logr.Logger, e exec.Interface) (*infrarepo.GitRemoteRepoInfraService, error) {
	gitCommandDriver, err := gitcommand.NewGitCommandDriver(l, e)
	if err != nil {
		return nil, err
	}
	gitRemoteRepoInfraService := infrarepo.NewGitRemoteRepoInfraService(gitCommandDriver, l)
	return gitRemoteRepoInfraService, nil
}
