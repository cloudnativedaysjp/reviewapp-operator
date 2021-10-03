//+build wireinject

package wire

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"k8s.io/utils/exec"

	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/gitapi"
	gitapi_iface "github.com/cloudnativedaysjp/reviewapp-operator/gateways/gitapi/iface"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/gitcommand"
	gitcommand_iface "github.com/cloudnativedaysjp/reviewapp-operator/gateways/gitcommand/iface"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/apprepo"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/infrarepo"
)

func NewGitRemoteRepoAppService(l logr.Logger) (*apprepo.GitRemoteRepoAppService, error) {
	wire.Build(
		gitapi.NewGitApiDriver,
		wire.Bind(new(gitapi_iface.GitApiIFace), new(*gitapi.GitApiDriver)),
		apprepo.NewGitRemoteRepoAppService,
	)
	return nil, nil
}

func NewGitRemoteRepoInfraService(l logr.Logger, e exec.Interface) (*infrarepo.GitRemoteRepoInfraService, error) {
	wire.Build(
		gitcommand.NewGitCommandDriver,
		wire.Bind(new(gitcommand_iface.GitCommandIFace), new(*gitcommand.GitCommandDriver)),
		infrarepo.NewGitRemoteRepoInfraService,
	)
	return nil, nil
}
