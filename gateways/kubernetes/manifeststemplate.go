package kubernetes

import (
	"context"
	"reflect"

	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

func (c Client) GetManifestsTemplate(ctx context.Context, m models.ReviewAppOrReviewAppManager) ([]models.ManifestsTemplate, error) {
	conf := m.GetInfraRepoConfig()
	var mts []models.ManifestsTemplate
	for _, tmp := range conf.Manifests.Templates {
		nn := types.NamespacedName(tmp)
		var mtOne dreamkastv1alpha1.ManifestsTemplate
		if err := c.Get(ctx, nn, &mtOne); err != nil {
			wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(mtOne), err)
			if apierrors.IsNotFound(err) {
				return nil, myerrors.NewK8sObjectNotFound(wrapedErr, mtOne.GVK(), nn)
			}
			return nil, wrapedErr
		}
		mtOne.SetGroupVersionKind(mtOne.GVK())
		mts = append(mts, models.ManifestsTemplate(mtOne))
	}
	return mts, nil
}
