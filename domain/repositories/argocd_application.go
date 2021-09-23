package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type ArgoCDApplictionIFace interface {
	GetAnnotationOfArgoCDApplication(ctx context.Context, namespace, name, annotationKey string) (commitHash string, err error)
	PrintArgoCDApplicationWithAnnotation(ctx context.Context, application models.ArgoCDApplicationString, annotationKey, annotationValue string) (models.ArgoCDApplicationString, error)
	PrintArgoCDApplicationName(ctx context.Context, application models.ArgoCDApplicationString) (string, error)
	PrintArgoCDApplicationNamespace(ctx context.Context, application models.ArgoCDApplicationString) (string, error)
}
