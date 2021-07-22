//+ wireinject

package wire

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/services"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/githubapi"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways/kubernetes"
)

func NewPullRequest(l logr.Logger, username string) (*repositories.PullRequest, error) {
	wire.Build(
		githubapi.NewGitHubApiGateway,
		repositories.NewPullRequest,
		wire.Bind(new(repositories.PullRequestIFace), new(*githubapi.GitHubApiGateway)),
	)
	return nil, nil
}

func NewPullRequestInfra(l logr.Logger, username string) (*repositories.PullRequestInfra, error) {
	wire.Build(
		githubapi.NewGitHubApiGateway,
		repositories.NewPullRequestInfra,
		wire.Bind(new(repositories.PullRequestInfraIFace), new(*githubapi.GitHubApiGateway)),
	)
	return nil, nil
}

func NewKubernetes(l logr.Logger, c client.Client) (*repositories.Kubernetes, error) {
	wire.Build(
		kubernetes.NewKubernetesGateway,
		repositories.NewKubernetes,
		wire.Bind(new(repositories.KubernetesIFace), new(*kubernetes.KubernetesGateway)),
	)
	return nil, nil
}

func NewGitRemoteRepoAppService(l logr.Logger, c client.Client, username string) (*services.GitRemoteRepoAppService, error) {
	wire.Build(
		githubapi.NewGitHubApiGateway,
		wire.Bind(new(repositories.PullRequestIFace), new(*githubapi.GitHubApiGateway)),
		kubernetes.NewKubernetesGateway,
		wire.Bind(new(repositories.GitRepoCredentialIFace), new(*kubernetes.KubernetesGateway)),
		services.NewGitRemoteRepoAppService,
	)
	return nil, nil
}

func NewGitRemoteRepoInfraService(l logr.Logger, c client.Client, username string) (*services.GitRemoteRepoInfraService, error) {
	wire.Build(
		githubapi.NewGitHubApiGateway,
		wire.Bind(new(repositories.PullRequestInfraIFace), new(*githubapi.GitHubApiGateway)),
		kubernetes.NewKubernetesGateway,
		wire.Bind(new(repositories.GitRepoCredentialIFace), new(*kubernetes.KubernetesGateway)),
		services.NewGitRemoteRepoInfraService,
	)
	return nil, nil
}

//repositories.PullRequestIFace, secret repositories.GitRepoCredentialIFace
