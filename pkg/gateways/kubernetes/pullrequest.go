package kubernetes

import (
	"context"
	"reflect"

	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cenkalti/backoff/v4"
	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/pkg/errors"
)

func (c Client) ApplyPullRequestWithOwnerRef(ctx context.Context,
	ra dreamkastv1alpha1.PullRequest, owner dreamkastv1alpha1.ReviewAppManager,
) error {
	if err := ctrl.SetControllerReference(&owner, &ra, c.Scheme()); err != nil {
		return err
	}
	return c.ApplyPullRequest(ctx, ra)
}

func (c Client) ApplyPullRequest(ctx context.Context,
	pr dreamkastv1alpha1.PullRequest,
) error {
	prApplied := &dreamkastv1alpha1.PullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pr.Name,
			Namespace: pr.Namespace,
		},
	}
	if _, err := ctrl.CreateOrUpdate(ctx, c, prApplied, func() (err error) {
		prApplied.Spec = pr.Spec
		prApplied.OwnerReferences = pr.OwnerReferences
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (c Client) PatchPullRequestStatus(ctx context.Context, pr dreamkastv1alpha1.PullRequest) error {
	// get curret PullRequest object (retry: 3)
	var prCurrent dreamkastv1alpha1.PullRequest
	if err := backoff.Retry(func() error {
		nn := types.NamespacedName{Name: pr.Name, Namespace: pr.Namespace}
		if err := c.Get(ctx, nn, &prCurrent); err != nil {
			wprpedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(prCurrent), err)
			if apierrors.IsNotFound(err) {
				return myerrors.NewK8sObjectNotFound(wprpedErr, prCurrent.GVK(), nn)
			}
			return wprpedErr
		}
		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)); err != nil {
		return err
	}
	// patch to PullRequest object
	patch := client.MergeFrom(&prCurrent)
	newPr := prCurrent.DeepCopy()
	newPr.Status = pr.Status
	if err := c.Status().Patch(ctx, newPr, patch); err != nil {
		// TODO
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(prCurrent), err)
	}
	return nil
}

func (c Client) GetPullRequest(ctx context.Context, namespace, name string) (dreamkastv1alpha1.PullRequest, error) {
	var pr dreamkastv1alpha1.PullRequest
	nn := types.NamespacedName{Namespace: namespace, Name: name}
	if err := c.Get(ctx, nn, &pr); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(pr), err)
		if apierrors.IsNotFound(err) {
			return dreamkastv1alpha1.PullRequest{}, myerrors.NewK8sObjectNotFound(wrapedErr, pr.GVK(), nn)
		}
		return dreamkastv1alpha1.PullRequest{}, wrapedErr
	}
	pr.SetGroupVersionKind(pr.GVK())
	return pr, nil
}

func (c Client) DeletePullRequest(ctx context.Context, namespace, name string) error {
	prDeleted := &dreamkastv1alpha1.PullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return c.Delete(ctx, prDeleted)
}
