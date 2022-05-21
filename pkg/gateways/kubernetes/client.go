package kubernetes

import (
	"context"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	dreamkastv1alpha1_iface "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1/iface"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	client.Client

	logger logr.Logger
}

func NewClient(l logr.Logger, c client.Client) *Client {
	return &Client{c, l}
}

type KubernetesIface interface {
	GetApplicationTemplate(context.Context, dreamkastv1alpha1.ReviewAppCommonSpec) (dreamkastv1alpha1.ApplicationTemplate, error)
	GetArgoCDAppFromReviewAppStatus(context.Context, dreamkastv1alpha1.ReviewAppStatus) (dreamkastv1alpha1.Application, error)
	GetLatestJobFromLabel(ctx context.Context, namespace, labelKey, labelValue string) (*batchv1.Job, error)
	CreateJob(context.Context, *batchv1.Job) error
	GetPreStopJobTemplate(ctx context.Context, ra dreamkastv1alpha1.ReviewApp) (dreamkastv1alpha1.JobTemplate, error)
	GetManifestsTemplate(context.Context, dreamkastv1alpha1.ReviewAppCommonSpec) ([]dreamkastv1alpha1.ManifestsTemplate, error)
	GetReviewApp(ctx context.Context, namespace, name string) (dreamkastv1alpha1.ReviewApp, error)
	ApplyReviewAppWithOwnerRef(ctx context.Context, ra dreamkastv1alpha1.ReviewApp, owner dreamkastv1alpha1.ReviewAppManager) error
	PatchReviewAppStatus(ctx context.Context, ra dreamkastv1alpha1.ReviewApp) error
	DeleteReviewApp(ctx context.Context, namespace, name string) error
	AddFinalizersToReviewApp(ctx context.Context, ra dreamkastv1alpha1.ReviewApp, finalizers ...string) error
	RemoveFinalizersFromReviewApp(ctx context.Context, ra dreamkastv1alpha1.ReviewApp, finalizers ...string) error
	GetReviewAppManager(ctx context.Context, namespace, name string) (dreamkastv1alpha1.ReviewAppManager, error)
	GetSecretValue(ctx context.Context, namespace string, m dreamkastv1alpha1_iface.AppOrInfraRepoTarget) (string, error)
	ApplyPullRequest(ctx context.Context, pr dreamkastv1alpha1.PullRequest) error
	ApplyPullRequestWithOwnerRef(ctx context.Context, pr dreamkastv1alpha1.PullRequest, owner dreamkastv1alpha1.ReviewAppManager) error
	PatchPullRequestStatus(ctx context.Context, pr dreamkastv1alpha1.PullRequest) error
	GetPullRequest(ctx context.Context, namespace, name string) (dreamkastv1alpha1.PullRequest, error)
	DeletePullRequest(ctx context.Context, namespace, name string) error
}

func (c Client) apply(ctx context.Context, obj *unstructured.Unstructured) error {
	return c.Patch(ctx, obj, client.Apply, &client.PatchOptions{
		FieldManager: "reviewapp-operator",
		Force:        pointer.Bool(true),
	})
}
