package services

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/repositories"
	"github.com/go-logr/logr"
)

type ReviewAppService struct {
	client.Client
	Log logr.Logger

	K8sFactory    repositories.KubernetesFactory
	gitApiFactory repositories.GitApiFactory
}

func NewReviewAppService(c client.Client, l logr.Logger, k8sFactory repositories.KubernetesFactory, gitApiFactory repositories.GitApiFactory) *ReviewAppService {
	return &ReviewAppService{c, l, k8sFactory, gitApiFactory}
}

func (s *ReviewAppService) ReconcileByPullRequest(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (ctrl.Result, error) {
	// init k8s repository for ReviewApp
	k8sRepo, err := s.K8sFactory.NewRepository(s.Client, s.Log)
	if err != nil {
		return reconcile.Result{}, err
	}

	// get Git AccessToken from Secret
	token, err := k8sRepo.GetSecretValue(ctx, types.NamespacedName{Name: ra.Spec.App.GitSecretRef.Name, Namespace: ra.Namespace}, ra.Spec.App.GitSecretRef.Key)
	if err != nil {
		return reconcile.Result{}, err
	}

	// init GitApi repository for ReviewApp
	gitapiRepo, err := s.gitApiFactory.NewRepository(s.Log, ra.Spec.App.Username, token)
	if err != nil {
		return reconcile.Result{}, err
	}

	// list PRs
	prs, err := gitapiRepo.ListPullRequestsWithOpen(ctx, ra.Spec.App.Organization, ra.Spec.App.Repository)
	if err != nil {
		return ctrl.Result{}, err
	}

	// apply ReviewAppInstance
	var syncedPullRequests []int
	for _, pr := range prs {
		// TODO: merge Template & generate ReviewAppInstance
		rai := &dreamkastv1beta1.ReviewAppInstance{}

		if err := k8sRepo.ApplyReviewAppInstanceFromReviewApp(
			ctx, types.NamespacedName{Name: ra.Name, Namespace: ra.Namespace},
			ra, rai,
		); err != nil {
			return ctrl.Result{}, err
		}

		syncedPullRequests = append(syncedPullRequests, pr.Number)
	}
	// TODO: PR が close 済な ReviewAppInstance を削除する
	// syncedPullRequests
	// ra.Status.SyncedPullRequests

	// update ReviewApp Status
	ra.Status.SyncedPullRequests = syncedPullRequests
	if err := k8sRepo.UpdateReviewAppStatus(ctx, ra); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
