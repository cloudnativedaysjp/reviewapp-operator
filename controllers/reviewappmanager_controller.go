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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/pkg/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/githubapi"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/kubernetes"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/metrics"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/template"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
)

// ReviewAppManagerReconciler reconciles a ReviewAppManager object
type ReviewAppManagerReconciler struct {
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	K8s    kubernetes.KubernetesIface
	GitApi githubapi.GitApiIface
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewappmanagers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewappmanagers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=applicationtemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=manifeststemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *ReviewAppManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("fetching ReviewAppManager resource: %s/%s", req.Namespace, req.Name))
	ram, err := r.K8s.GetReviewAppManager(ctx, req.Namespace, req.Name)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.removeMetrics(req.Name, req.Namespace)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, ram)
}

func (r *ReviewAppManagerReconciler) reconcile(ctx context.Context, ram dreamkastv1alpha1.ReviewAppManager) (ctrl.Result, error) {
	appRepoTarget := ram.Spec.AppTarget

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoToken, err := r.K8s.GetSecretValue(ctx, ram.Namespace, &appRepoTarget)
	if err != nil {
		if myerrors.IsNotFound(err) || myerrors.IsKeyMissing(err) {
			r.Log.Info(err.Error())
			return defaultResult, nil
		}
		return defaultResult, err
	}
	// set credential
	gitRemoteRepoCred := models.NewGitCredential(ram.Spec.AppTarget.Username, gitRemoteRepoToken)
	if err := r.GitApi.WithCredential(gitRemoteRepoCred); err != nil {
		return defaultResult, err
	}
	// list PRs
	prs, err := r.GitApi.ListOpenPullRequests(ctx, appRepoTarget)
	if err != nil {
		return defaultResult, err
	}
	prs = prs.ExcludeSpecificPR(ram.Spec.AppTarget) // exclude PRs with specific labels
	// add metrics
	metrics.RequestToGitHubApiCounterVec.WithLabelValues(
		ram.Name,
		ram.Namespace,
		"ReviewAppManager",
	).Add(1)

	for _, pr := range prs.Items {
		pr.Namespace = ram.Namespace
		v := template.NewTemplator(ram.Spec.ReviewAppCommonSpec, pr)
		// generate RA
		ra, err := v.ReviewApp(ram, pr)
		if err != nil {
			return defaultResult, err
		}
		// apply RA
		if err := r.K8s.ApplyReviewAppWithOwnerRef(ctx, ra, ram); err != nil {
			return defaultResult, err
		}
		// apply PR
		if err := r.K8s.ApplyPullRequestWithOwnerRef(ctx, pr, ram); err != nil {
			return defaultResult, err
		}
	}

	return defaultResult, nil
}

func (r *ReviewAppManagerReconciler) removeMetrics(name, namespace string) {
	metrics.RequestToGitHubApiCounterVec.DeleteLabelValues(
		name,
		namespace,
		"ReviewAppManager",
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReviewAppManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&dreamkastv1alpha1.ReviewAppManager{}).
		Owns(&dreamkastv1alpha1.ReviewApp{}).
		Owns(&dreamkastv1alpha1.PullRequest{}).
		Complete(r)
}
