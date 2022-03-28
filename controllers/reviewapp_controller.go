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

	"github.com/go-logr/logr"
	"golang.org/x/sync/singleflight"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/metrics"
)

// ReviewAppReconciler reconciles a ReviewApp object
type ReviewAppReconciler struct {
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	K8sRepository        repositories.KubernetesRepository
	GitApiRepository     repositories.GitAPI
	GitCommandRepository repositories.GitCommand
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch

const (
	finalizer = "reviewapp.finalizers.cloudnativedays.jp"
)

var (
	singleflightGroupForReviewApp singleflight.Group
)

func (r *ReviewAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	result, err, _ := singleflightGroupForReviewApp.Do(fmt.Sprintf("%s/%s", req.Namespace, req.Name), func() (interface{}, error) {
		r.Log.Info(fmt.Sprintf("fetching ReviewApp resource: %s/%s", req.Namespace, req.Name))
		ra, err := r.K8sRepository.GetReviewApp(ctx, req.Namespace, req.Name)
		if err != nil {
			if myerrors.IsNotFound(err) {
				r.Log.Info(fmt.Sprintf("%s %s/%s not found", reflect.TypeOf(ra), req.Namespace, req.Name))
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		dto, result, err := r.prepare(ctx, ra)
		if err != nil {
			if myerrors.IsNotFound(err) {
				return result, nil
			}
			return result, err
		}

		// Add Finalizers
		if err := r.K8sRepository.AddFinalizersToReviewApp(ctx, ra, finalizer); err != nil {
			return ctrl.Result{}, err
		}

		// Handle deletion reconciliation loop.
		if !ra.ObjectMeta.DeletionTimestamp.IsZero() {
			return r.reconcileDelete(ctx, *dto)
		}
		return r.reconcile(ctx, *dto)
	})
	return result.(ctrl.Result), err
}

func (r *ReviewAppReconciler) reconcile(ctx context.Context, dto ReviewAppPhaseDTO) (result ctrl.Result, err error) {
	ra := dto.ReviewApp
	metrics.SetMetricsUp(dto.ReviewApp)

	// run/skip processes by ReviewApp state
	errs := []error{}
	if reflect.DeepEqual(ra.Status, dreamkastv1alpha1.ReviewAppStatus{}) ||
		ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeWatchingAppRepo {
		ra, result, err = r.confirmUpdated(ctx, dto)
		if err != nil {
			errs = append(errs, err)
		}
		dto.ReviewApp = ra
	}
	if ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo {
		ra, result, err = r.deployReviewAppManifestsToInfraRepo(ctx, dto)
		if err != nil {
			errs = append(errs, err)
		}
		dto.ReviewApp = ra
	}
	if ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeUpdatedInfraRepo {
		ra, result, err = r.commentToAppRepoPullRequest(ctx, dto)
		if err != nil {
			errs = append(errs, err)
		}
		dto.ReviewApp = ra
	}

	// update status
	if err := r.K8sRepository.UpdateReviewAppStatus(ctx, ra); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, kerrors.NewAggregate(errs)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReviewAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dreamkastv1alpha1.ReviewApp{}).
		Complete(r)
}
