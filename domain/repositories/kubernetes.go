package repositories

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type KubernetesRepository interface {
	GetApplicationTemplate(ctx context.Context, m models.ReviewAppOrReviewAppManager) (models.ApplicationTemplate, error)
	GetArgoCDAppFromReviewAppStatus(ctx context.Context, raStatus models.ReviewAppStatus) (models.Application, error)
	GetLatestJobFromLabel(ctx context.Context, namespace, labelKey, labelValue string) (*batchv1.Job, error)
	CreateJob(ctx context.Context, job *batchv1.Job) error
	GetPreStopJobTemplate(ctx context.Context, ra models.ReviewApp) (models.JobTemplate, error)
	GetManifestsTemplate(ctx context.Context, m models.ReviewAppOrReviewAppManager) ([]models.ManifestsTemplate, error)
	GetReviewApp(ctx context.Context, namespace, name string) (models.ReviewApp, error)
	ApplyReviewAppWithOwnerRef(ctx context.Context, ra models.ReviewApp, owner models.ReviewAppManager) error
	ApplyReviewAppStatus(ctx context.Context, ra models.ReviewApp) error
	DeleteReviewApp(ctx context.Context, namespace, name string) error
	AddFinalizersToReviewApp(ctx context.Context, ra models.ReviewApp, finalizers ...string) error
	RemoveFinalizersFromReviewApp(ctx context.Context, ra models.ReviewApp, finalizers ...string) error
	GetReviewAppManager(ctx context.Context, namespace, name string) (models.ReviewAppManager, error)
	UpdateReviewAppManagerStatus(ctx context.Context, ram models.ReviewAppManager) error
	GetSecretValue(ctx context.Context, namespace string, m models.AppOrInfraRepoTarget) (string, error)
}
