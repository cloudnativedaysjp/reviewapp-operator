package repositories

import (
	"context"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type KubernetesRepository interface {
	GetApplicationTemplate(ctx context.Context, m models.ReviewAppOrReviewAppManager) (models.ApplicationTemplate, error)
	GetArgoCDAppFromReviewAppStatus(ctx context.Context, ra models.ReviewApp) (models.Application, error)
	GetLatestJobFromLabel(ctx context.Context, namespace, labelKey, labelValue string) (*batchv1.Job, error)
	CreateJob(ctx context.Context, job *batchv1.Job) error
	GetPreStopJobTemplate(ctx context.Context, ra models.ReviewApp) (models.JobTemplate, error)
	GetManifestsTemplate(ctx context.Context, m models.ReviewAppOrReviewAppManager) ([]models.ManifestsTemplate, error)
	GetReviewApp(ctx context.Context, namespace, name string) (*dreamkastv1alpha1.ReviewApp, error)
	ApplyReviewAppWithOwnerRef(ctx context.Context, ra models.ReviewApp, owner metav1.Object) error
	UpdateReviewAppStatus(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp) error
	DeleteReviewApp(ctx context.Context, namespace, name string) error
	AddFinalizersToReviewApp(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp, finalizers ...string) error
	RemoveFinalizersFromReviewApp(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp, finalizers ...string) error
	GetReviewAppManager(ctx context.Context, namespace, name string) (*dreamkastv1alpha1.ReviewAppManager, error)
	UpdateReviewAppManagerStatus(ctx context.Context, ram models.ReviewAppManager) error
	GetSecretValue(ctx context.Context, namespace string, m models.AppOrInfraRepoTarget) (string, error)
}
