/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/pkg/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/models"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/githubapi"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/kubernetes"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/metrics"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
)

// PullRequestReconciler reconciles a PullRequest object
type PullRequestReconciler struct {
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	K8s    kubernetes.KubernetesIface
	GitApi githubapi.GitApiIface
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=pullrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=pullrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=pullrequests/finalizers,verbs=update

func (r *PullRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("fetching PullRequest resource: %s/%s", req.Namespace, req.Name))
	pr, err := r.K8s.GetPullRequest(ctx, req.Namespace, req.Name)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.removeMetrics(req.Name, req.Namespace)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, pr)
}

func (r *PullRequestReconciler) reconcile(ctx context.Context, pr dreamkastv1alpha1.PullRequest) (ctrl.Result, error) {
	// get gitRemoteRepo credential from Secret
	gitRemoteRepoToken, err := r.K8s.GetSecretValue(ctx, pr.Namespace, pr.Spec.AppTarget)
	if err != nil {
		if myerrors.IsNotFound(err) || myerrors.IsKeyMissing(err) {
			r.Log.Info(err.Error())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	gitRemoteRepoCred := models.NewGitCredential(pr.Spec.AppTarget.Username, gitRemoteRepoToken)
	if err := r.GitApi.WithCredential(gitRemoteRepoCred); err != nil {
		return ctrl.Result{}, err
	}
	// Get PullRequest by GitHub API.
	// If PR already closed, delete PR object myself.
	// (concidered GitHub API's error, retry 3 times)
	var prFromGitApi dreamkastv1alpha1.PullRequest
	if err := backoff.Retry(func() error {
		prFromGitApi, err = r.GitApi.GetPullRequest(ctx, pr.Spec.AppTarget, pr.Spec.Number)
		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)); err != nil {
		if err := r.K8s.DeletePullRequest(ctx, pr.Namespace, pr.Name); err != nil {
			return defaultResult, err
		}
		return ctrl.Result{}, nil
	}
	// add metrics
	metrics.RequestToGitHubApiCounterVec.WithLabelValues(
		pr.Name,
		pr.Namespace,
		"PullRequest",
	).Add(1)

	// delete PullRequest object if PR matched ignore conditions
	if pr.MustBeIgnored(pr.Spec.AppTarget) {
		if err := r.K8s.DeletePullRequest(ctx, pr.Namespace, pr.Name); err != nil {
			return defaultResult, err
		}
		return ctrl.Result{}, nil
	}

	// apply PullRequest to K8s
	if pr.Name != prFromGitApi.Name {
		return defaultResult, fmt.Errorf("TODO")
	}
	pr.Status = prFromGitApi.Status
	if err := r.K8s.PatchPullRequestStatus(ctx, pr); err != nil {
		return defaultResult, err
	}
	return ctrl.Result{
		RequeueAfter: 30 * time.Second,
	}, nil
}

func (r *PullRequestReconciler) removeMetrics(name, namespace string) {
	metrics.RequestToGitHubApiCounterVec.DeleteLabelValues(
		name,
		namespace,
		"PullRequest",
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PullRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	setupLog := ctrl.Log.WithName("setup")
	var err error
	r.K8s, err = wire.NewKubernetes(r.Log, mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "unable to initialize", "wire.NewKubernetesRepository")
		os.Exit(1)
	}
	r.GitApi, err = wire.NewGitHubApi(r.Log)
	if err != nil {
		setupLog.Error(err, "unable to initialize", "wire.NewGitHubAPIRepository")
		os.Exit(1)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&dreamkastv1alpha1.PullRequest{}).
		Complete(r)
}
