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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/apprepo"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/kubernetes"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/template"
)

// ReviewAppManagerReconciler reconciles a ReviewAppManager object
type ReviewAppManagerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	GitRemoteRepoAppService *apprepo.GitRemoteRepoAppService
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewappmanagers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewappmanagers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewappmanagers/finalizers,verbs=update

func (r *ReviewAppManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var ram *dreamkastv1beta1.ReviewAppManager
	r.Log.Info(fmt.Sprintf("fetching %s resource", reflect.TypeOf(ram)))
	ram, err := kubernetes.GetReviewAppManager(ctx, r.Client, req.Namespace, req.Name)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("%s %s/%s not found", reflect.TypeOf(ram), req.Namespace, req.Name))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, ram)
}

func (r *ReviewAppManagerReconciler) reconcile(ctx context.Context, ram *dreamkastv1beta1.ReviewAppManager) (ctrl.Result, error) {
	// get gitRemoteRepo credential from Secret
	gitRemoteRepoCred, err := kubernetes.GetSecretValue(ctx,
		r.Client, ram.Namespace, ram.Spec.AppTarget.GitSecretRef.Name, ram.Spec.AppTarget.GitSecretRef.Key,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// list PRs
	prs, err := r.GitRemoteRepoAppService.ListOpenPullRequest(ctx,
		ram.Spec.AppTarget.Organization, ram.Spec.AppTarget.Repository,
		ram.Spec.AppTarget.Username, gitRemoteRepoCred,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// apply ReviewApp
	var syncedPullRequests []dreamkastv1beta1.ReviewAppManagerStatusSyncedPullRequests
	for _, pr := range prs {

		// if PR labeled with models.CandidateLabelName, using candidate template in ApplicationTemplate / ManifestsTemplate
		isCandidate := false
		for _, l := range pr.Labels {
			if l == models.CandidateLabelName {
				isCandidate = true
			}
		}

		// generate RA struct
		ra := kubernetes.NewReviewAppFromReviewAppManager(ram, pr)
		ra.Spec.AppTarget = ram.Spec.AppTarget
		ra.Spec.InfraTarget = ram.Spec.InfraTarget

		// Templating
		{
			v := template.NewTemplateValue(
				pr.Organization, pr.Repository, pr.Number, pr.HeadCommitSha,
				ram.Spec.InfraTarget.Organization, ram.Spec.InfraTarget.Repository, ra.Status.Sync.InfraRepoLatestCommitSha,
				kubernetes.PickVariablesFromReviewAppManager(ctx, ram),
			)
			{ // template from ram.Spec.AppConfig to ra.Spec.AppConfig
				out, err := yaml.Marshal(&ram.Spec.AppConfig)
				if err != nil {
					return ctrl.Result{}, err
				}
				appConfigStr, err := v.Templating(string(out))
				if err != nil {
					return ctrl.Result{}, err
				}
				if err := yaml.Unmarshal([]byte(appConfigStr), &ra.Spec.AppConfig); err != nil {
					return ctrl.Result{}, err
				}
			}
			{ // template from ram.Spec.InfraConfig to ra.Spec.InfraConfig
				out, err := yaml.Marshal(&ram.Spec.InfraConfig)
				if err != nil {
					return ctrl.Result{}, err
				}
				infraConfigStr, err := v.Templating(string(out))
				if err != nil {
					return ctrl.Result{}, err
				}
				if err := yaml.Unmarshal([]byte(infraConfigStr), &ra.Spec.InfraConfig); err != nil {
					return ctrl.Result{}, err
				}
			}
			{ // get ApplicationTemplate & template to ra.Spec.Application
				at, err := kubernetes.GetApplicationTemplate(ctx, r.Client, ram.Spec.InfraConfig.ArgoCDApp.Template.Namespace, ram.Spec.InfraConfig.ArgoCDApp.Template.Name)
				if err != nil {
					if myerrors.IsNotFound(err) {
						r.Log.Info(fmt.Sprintf("%s %s/%s not found", reflect.TypeOf(at), ram.Namespace, ram.Name))
						return ctrl.Result{}, nil
					}
					return ctrl.Result{}, err
				}
				if isCandidate {
					ra.Spec.Application, err = v.Templating(at.Spec.CandidateTemplate)
				} else {
					ra.Spec.Application, err = v.Templating(at.Spec.StableTemplate)
				}
				if err != nil {
					return ctrl.Result{}, err
				}
			}
			{ // get ManifestsTemplate & template to ra.Spec.Manifests
				for _, mt := range ram.Spec.InfraConfig.Manifests.Templates {
					mt, err := kubernetes.GetManifestsTemplate(ctx, r.Client, mt.Namespace, mt.Name)
					if err != nil {
						if myerrors.IsNotFound(err) {
							r.Log.Info(fmt.Sprintf("%s %s/%s not found", reflect.TypeOf(mt), ram.Namespace, ram.Name))
							return ctrl.Result{}, nil
						}
						return ctrl.Result{}, err
					}
					if isCandidate {
						ra.Spec.Manifests, err = v.MapTemplatingAndAppend(ra.Spec.Manifests, mt.Spec.CandidateData)
					} else {
						ra.Spec.Manifests, err = v.MapTemplatingAndAppend(ra.Spec.Manifests, mt.Spec.StableData)
					}
					if err != nil {
						return ctrl.Result{}, err
					}
				}
			}
		}

		// apply RA
		if err := kubernetes.ApplyReviewAppWithOwnerRef(ctx, r.Client, ra, ram); err != nil {
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
		if err := kubernetes.DeleteReviewApp(ctx, r.Client, ram.Namespace, a.ReviewAppName); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	// update ReviewAppManager Status
	ram.Status.SyncedPullRequests = syncedPullRequests
	if err := kubernetes.UpdateReviewAppManagerStatus(ctx, r.Client, ram); err != nil {
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
