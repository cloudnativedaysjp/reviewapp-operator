package services

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	k8sRepo := s.K8sFactory.New(s)
	gitapiRepo := s.K8sFactory.New(s)

	// list PRs
	prs, err := gitapiRepo.ListPullRequests(project, repo)
	if err != nil {
		return ctrl.Result{}, err
	}

	// apply ReviewAppInstance
	for pr := range prs {
		k8sRepo.ApplyReviewAppInstance(TODO)
	}

	return ctrl.Result{}, nil
}
