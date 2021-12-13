package kubernetes

import (
	"context"
	"reflect"
	"strings"

	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

func GetReviewAppManager(ctx context.Context, c client.Client, namespace, name string) (*dreamkastv1alpha1.ReviewAppManager, error) {
	var ram dreamkastv1alpha1.ReviewAppManager
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &ram); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(ram), err)
		if apierrors.IsNotFound(err) {
			return nil, myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return nil, wrapedErr
	}
	return &ram, nil
}

func UpdateReviewAppManagerStatus(ctx context.Context, c client.Client, ram *dreamkastv1alpha1.ReviewAppManager) error {
	var ramCurrent dreamkastv1alpha1.ReviewAppManager
	if err := c.Get(ctx, types.NamespacedName{Name: ram.Name, Namespace: ram.Namespace}, &ramCurrent); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(ramCurrent), err)
		if apierrors.IsNotFound(err) {
			return myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return wrapedErr
	}

	ramCurrent.Status = ram.Status
	if err := c.Status().Update(ctx, &ramCurrent); err != nil {
		return xerrors.Errorf("Error to Update %s: %w", reflect.TypeOf(ramCurrent), err)
	}
	return nil
}

func PickVariablesFromReviewAppManager(ctx context.Context, ram *dreamkastv1alpha1.ReviewAppManager) map[string]string {
	vars := make(map[string]string)
	for _, line := range ram.Spec.Variables {
		idx := strings.Index(line, "=")
		if idx == -1 {
			// TODO
			// r.Log.Info(fmt.Sprintf("RA %s: .Spec.Variables[%d] is invalid", ram.Name, i))
			continue
		}
		vars[line[:idx]] = line[idx+1:]
	}
	return vars
}
