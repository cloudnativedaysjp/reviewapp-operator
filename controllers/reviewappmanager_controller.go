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
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/services"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
)

// ReviewAppManagerReconciler reconciles a ReviewAppManager object
type ReviewAppManagerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewappmanagers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewappmanagers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewappmanagers/finalizers,verbs=update

func (r *ReviewAppManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var ram dreamkastv1beta1.ReviewAppManager
	r.Log.Info(fmt.Sprintf("fetching %s resource", reflect.TypeOf(ram)))
	if err := r.Get(ctx, req.NamespacedName, &ram); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("%s %s/%s not found", reflect.TypeOf(ram), req.Namespace, req.Name))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, &ram)
}

func (r *ReviewAppManagerReconciler) reconcile(ctx context.Context, ram *dreamkastv1beta1.ReviewAppManager) (ctrl.Result, error) {
	gitRemoteRepoAppService, err := wire.NewGitRemoteRepoAppService(r.Log, r.Client, ram.Spec.App.Username)
	if err != nil {
		return ctrl.Result{}, err
	}
	k8sService, err := wire.NewKubernetesService(r.Log, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// list PRs
	prs, err := gitRemoteRepoAppService.ListOpenPullRequest(ctx,
		ram.Spec.App.Organization, ram.Spec.App.Repository,
		services.AccessToAppRepoInput{
			SecretNamespace: ram.Namespace,
			SecretName:      ram.Spec.App.GitSecretRef.Name,
			SecretKey:       ram.Spec.App.GitSecretRef.Key,
		},
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// apply ReviewApp
	var syncedPullRequests []dreamkastv1beta1.ReviewAppManagerStatusSyncedPullRequests
	for _, pr := range prs {
		// generate RA struct
		ra := &dreamkastv1beta1.ReviewApp{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s-%s-%d",
					ram.Name,
					strings.ToLower(pr.Organization),
					strings.ToLower(pr.Repository),
					pr.Number,
				),
				Namespace: ram.Namespace,
			},
		}

		// merge Template & generate ReviewApp
		if err := k8sService.MergeTemplate(ctx, ra, ram, pr); err != nil {
			if models.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}

		// apply RA
		if err := k8sService.ApplyReviewAppFromReviewAppManager(ctx, ra, ram); err != nil {
			return ctrl.Result{}, err
		}

		// set values for update status
		syncedPullRequests = append(syncedPullRequests, dreamkastv1beta1.ReviewAppManagerStatusSyncedPullRequests{
			Organization:  pr.Organization,
			Repository:    pr.Repository,
			Number:        pr.Number,
			ReviewAppName: ra.Name,
		})
	}

	// delete the ReviewApp that associated PR has already been closed
loop:
	for _, a := range ram.Status.SyncedPullRequests {
		for _, b := range syncedPullRequests {
			if a.Organization == b.Organization && a.Repository == b.Repository && a.Number == b.Number {
				continue loop
			}
		}
		// delete RA that only exists ResourceStatus
		if err := k8sService.ReviewAppIFace.DeleteReviewApp(ctx, ram.Namespace, a.ReviewAppName); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	// update ReviewAppManager Status
	ram.Status.SyncedPullRequests = syncedPullRequests
	if err := k8sService.UpdateReviewAppManagerStatus(ctx, ram); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReviewAppManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dreamkastv1beta1.ReviewAppManager{}).
		Owns(&dreamkastv1beta1.ReviewApp{}).
		// TODO: at, mt 更新時にも reconcile が走るようにする
		// Watches(
		// 	&source.Kind{Type: &dreamkastv1beta1.ApplicationTemplate{}},
		// 	&handler.EnqueueRequestForObject{},
		// ).
		// Watches(
		// 	&source.Kind{Type: &dreamkastv1beta1.ManifestsTemplate{}},
		// 	&handler.EnqueueRequestForObject{},
		// ).
		Complete(r)
}
