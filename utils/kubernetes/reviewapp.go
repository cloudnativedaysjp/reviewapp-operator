package kubernetes

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/models"
)

func NewReviewAppFromReviewAppManager(ram *dreamkastv1beta1.ReviewAppManager, pr *models.PullRequest) *dreamkastv1beta1.ReviewApp {
	return &dreamkastv1beta1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-%s-%d",
				ram.Name,
				strings.ToLower(pr.Organization),
				strings.ToLower(pr.Repository),
				pr.Number,
			),
			Namespace: ram.Namespace,
		},
		Spec: dreamkastv1beta1.ReviewAppSpec{
			AppTarget:   ram.Spec.AppTarget,
			InfraTarget: ram.Spec.InfraTarget,
			AppPrNum:    pr.Number,
		},
	}
}

func GetReviewApp(ctx context.Context, c client.Client, namespace, name string) (*dreamkastv1beta1.ReviewApp, error) {
	var ra dreamkastv1beta1.ReviewApp
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &ra); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(ra), err)
		if apierrors.IsNotFound(err) {
			return nil, myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return nil, wrapedErr
	}
	return &ra, nil
}

func ApplyReviewAppWithOwnerRef(ctx context.Context, c client.Client, ra *dreamkastv1beta1.ReviewApp, owner metav1.Object) error {
	raApplied := &dreamkastv1beta1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ra.Name,
			Namespace: ra.Namespace,
		},
	}
	if _, err := ctrl.CreateOrUpdate(ctx, c, raApplied, func() (err error) {
		raApplied.Spec = ra.Spec
		if err := ctrl.SetControllerReference(owner, raApplied, c.Scheme()); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func UpdateReviewAppStatus(ctx context.Context, c client.Client, ra *dreamkastv1beta1.ReviewApp) error {
	var raCurrent dreamkastv1beta1.ReviewApp
	if err := c.Get(ctx, types.NamespacedName{Name: ra.Name, Namespace: ra.Namespace}, &raCurrent); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(raCurrent), err)
		if apierrors.IsNotFound(err) {
			return myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return wrapedErr
	}
	patch := client.MergeFrom(&raCurrent)

	if err := c.Status().Patch(ctx, ra, patch); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(raCurrent), err)
	}
	return nil
}

func DeleteReviewApp(ctx context.Context, c client.Client, namespace, name string) error {
	ra := dreamkastv1beta1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := c.Delete(ctx, &ra); err != nil {
		return xerrors.Errorf("Error to Delete %s: %w", reflect.TypeOf(ra), err)
	}
	return nil
}

func AddFinalizersToReviewApp(ctx context.Context, c client.Client, ra *dreamkastv1beta1.ReviewApp, finalizers ...string) error {
	raPatched := *ra.DeepCopy()
	for _, f := range finalizers {
		controllerutil.AddFinalizer(&raPatched, f)
	}
	patch := client.MergeFrom(ra)
	if err := c.Patch(ctx, &raPatched, patch); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(raPatched), err)
	}
	return nil
}

func RemoveFinalizersToReviewApp(ctx context.Context, c client.Client, ra *dreamkastv1beta1.ReviewApp, finalizers ...string) error {
	raPatched := *ra.DeepCopy()
	for _, f := range finalizers {
		if controllerutil.ContainsFinalizer(&raPatched, f) {
			controllerutil.RemoveFinalizer(&raPatched, f)
		}
	}
	patch := client.MergeFrom(ra)
	if err := c.Patch(ctx, &raPatched, patch); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(raPatched), err)
	}
	return nil
}
