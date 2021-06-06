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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
)

// ReviewAppReconciler reconciles a ReviewApp object
type ReviewAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ReviewApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ReviewAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// list PRs

	// for PR := range PRs

	//// diff between PR.prNum and ReviewAppStatus.syncedArtifacts.appRepoPrNum
	//// if no diff then continue

	//// if new PR.prNum exists

	////// fetch ReviewApp.infra.repository
	////// create Application manifest from ApplicationTemplate
	////// create manifests from ManifestsTemplate
	////// push

	////// set appRepoPrNum, applicationName, appRepoSha, infraRepoSha to ReviewAppStatus

	//// elif less PR.prNum than ReviewAppStatus.syncedArtifacts.appRepoPrNum

	////// fetch ReviewApp.infra.repository
	////// delete Application manifest
	////// delete manifests
	////// push

	////// delete element from ReviewAppStatus

	//// endif

	// endfor

	// for artifact := range ReviewAppStatus.syncedArtifacts

	//// status := check ArgoCD Applications Synced Status
	//// if status.Sha == ReviewAppStatus.syncedArtifacts[].infraRepoSha && !ReviewAppStatus.syncedArtifacts[].notified

	////// notify to PR in app repo
	////// set ReviewAppStatus.syncedArtifacts[].notified = true

	//// endif

	// endfor

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReviewAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dreamkastv1beta1.ReviewApp{}).
		Watches(
			&source.Kind{Type: &dreamkastv1beta1.ApplicationTemplate{}},
			&handler.EnqueueRequestForObject{},
		).
		Watches(
			&source.Kind{Type: &dreamkastv1beta1.ManifestsTemplate{}},
			&handler.EnqueueRequestForObject{},
		).
		Complete(r)
}
