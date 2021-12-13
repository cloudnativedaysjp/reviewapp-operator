package kubernetes

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

type PullRequest struct {
	Organization string
	Repository   string
	Number       int
}

func NewReviewAppFromReviewAppManager(ram *dreamkastv1alpha1.ReviewAppManager, pr *PullRequest) *dreamkastv1alpha1.ReviewApp {
	return &dreamkastv1alpha1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-%s-%d",
				ram.Name,
				strings.ToLower(pr.Organization),
				strings.ToLower(pr.Repository),
				pr.Number,
			),
			Namespace: ram.Namespace,
		},
		Spec: dreamkastv1alpha1.ReviewAppSpec{
			AppTarget:   ram.Spec.AppTarget,
			InfraTarget: ram.Spec.InfraTarget,
			AppPrNum:    pr.Number,
		},
	}
}

func GetReviewApp(ctx context.Context, c client.Client, namespace, name string) (*dreamkastv1alpha1.ReviewApp, error) {
	var ra dreamkastv1alpha1.ReviewApp
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &ra); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(ra), err)
		if apierrors.IsNotFound(err) {
			return nil, myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return nil, wrapedErr
	}
	return &ra, nil
}

func ApplyReviewAppWithOwnerRef(ctx context.Context, c client.Client, ra *dreamkastv1alpha1.ReviewApp, owner metav1.Object) error {
	raApplied := &dreamkastv1alpha1.ReviewApp{
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

func UpdateReviewAppStatus(ctx context.Context, c client.Client, ra *dreamkastv1alpha1.ReviewApp) error {
	var raCurrent dreamkastv1alpha1.ReviewApp
	if err := c.Get(ctx, types.NamespacedName{Name: ra.Name, Namespace: ra.Namespace}, &raCurrent); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(raCurrent), err)
		if apierrors.IsNotFound(err) {
			return myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return wrapedErr
	}

	raCurrent.Status = ra.Status
	if err := c.Status().Update(ctx, &raCurrent); err != nil {
		return xerrors.Errorf("Error to Update %s: %w", reflect.TypeOf(raCurrent), err)
	}
	return nil
}

func DeleteReviewApp(ctx context.Context, c client.Client, namespace, name string) error {
	ra := dreamkastv1alpha1.ReviewApp{
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

func AddFinalizersToReviewApp(ctx context.Context, c client.Client, ra *dreamkastv1alpha1.ReviewApp, finalizers ...string) error {
	patch := &unstructured.Unstructured{}
	patch.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "dreamkast.cloudnativedays.jp",
		Version: "v1alpha1",
		Kind:    "ReviewApp",
	})
	patch.SetNamespace(ra.Namespace)
	patch.SetName(ra.Name)
	patch.SetFinalizers(ra.Finalizers)
	for _, f := range finalizers {
		controllerutil.AddFinalizer(patch, f)
	}

	if err := c.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "reviewapp-operator",
		Force:        pointer.Bool(true),
	}); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(ra), err)
	}
	return nil
}

func RemoveFinalizersToReviewApp(ctx context.Context, c client.Client, ra *dreamkastv1alpha1.ReviewApp, finalizers ...string) error {
	patch := &unstructured.Unstructured{}
	patch.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "dreamkast.cloudnativedays.jp",
		Version: "v1alpha1",
		Kind:    "ReviewApp",
	})
	patch.SetNamespace(ra.Namespace)
	patch.SetName(ra.Name)
	patch.SetFinalizers(ra.Finalizers)
	for _, f := range finalizers {
		if controllerutil.ContainsFinalizer(patch, f) {
			controllerutil.RemoveFinalizer(patch, f)
		}
	}
	if err := c.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "reviewapp-operator",
		Force:        pointer.Bool(true),
	}); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(ra), err)
	}
	return nil
}

func PickVariablesFromReviewApp(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp) map[string]string {
	vars := make(map[string]string)
	for _, line := range ra.Spec.Variables {
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
