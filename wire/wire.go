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

func NewArgoCDApplication(l logr.Logger, c client.Client) (*repositories.ArgoCDApplication, error) {
	wire.Build(
		kubernetes.NewKubernetesGateway,
		repositories.NewArgoCDApplication,
		wire.Bind(new(repositories.ArgoCDApplicationIFace), new(*kubernetes.KubernetesGateway)),
	)
	return nil, nil
}

func NewGitRepoSecret(l logr.Logger, c client.Client) (*repositories.GitRepoSecret, error) {
	wire.Build(
		kubernetes.NewKubernetesGateway,
		repositories.NewGitRepoSecret,
		wire.Bind(new(repositories.GitRepoSecretIFace), new(*kubernetes.KubernetesGateway)),
	)
	return nil, nil
}

func NewPullRequestApp(l logr.Logger, username string) (*repositories.PullRequestApp, error) {
	wire.Build(
		githubapi.NewGitHubApiGateway,
		repositories.NewPullRequestApp,
		wire.Bind(new(repositories.PullRequestAppIFace), new(*githubapi.GitHubApiGateway)),
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

func NewReviewAppConfig(l logr.Logger, c client.Client) (*repositories.ReviewAppConfig, error) {
	wire.Build(
		kubernetes.NewKubernetesGateway,
		repositories.NewReviewAppConfig,
		wire.Bind(new(repositories.ReviewAppConfigIFace), new(*kubernetes.KubernetesGateway)),
	)
	return nil, nil
}

func NewReviewAppInstance(l logr.Logger, c client.Client) (*repositories.ReviewAppInstance, error) {
	wire.Build(
		kubernetes.NewKubernetesGateway,
		repositories.NewReviewAppInstance,
		wire.Bind(new(repositories.ReviewAppInstanceIFace), new(*kubernetes.KubernetesGateway)),
	)
	return nil, nil
}

func NewGitRemoteRepoAppService(l logr.Logger, c client.Client, username string) (*services.GitRemoteRepoAppService, error) {
	wire.Build(
		githubapi.NewGitHubApiGateway,
		wire.Bind(new(repositories.PullRequestAppIFace), new(*githubapi.GitHubApiGateway)),
		kubernetes.NewKubernetesGateway,
		wire.Bind(new(repositories.GitRepoSecretIFace), new(*kubernetes.KubernetesGateway)),
		services.NewGitRemoteRepoAppService,
	)
	return nil, nil
}

func NewGitRemoteRepoInfraService(l logr.Logger, c client.Client, username string) (*services.GitRemoteRepoInfraService, error) {
	wire.Build(
		githubapi.NewGitHubApiGateway,
		wire.Bind(new(repositories.PullRequestInfraIFace), new(*githubapi.GitHubApiGateway)),
		kubernetes.NewKubernetesGateway,
		wire.Bind(new(repositories.GitRepoSecretIFace), new(*kubernetes.KubernetesGateway)),
		services.NewGitRemoteRepoInfraService,
	)
	return nil, nil
}

//repositories.PullRequestIFace, secret repositories.GitRepoSecretIFace
