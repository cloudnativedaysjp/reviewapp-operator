package services

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

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
	var syncedPullRequests []dreamkastv1beta1.ReviewAppStatusSyncedPullRequests
	for _, pr := range prs {
		// generate RAI struct
		rai := &dreamkastv1beta1.ReviewAppInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", ra.Name, pr.Number),
				Namespace: ra.Namespace,
			},
		}

		// merge Template & generate ReviewAppInstance
		if err := s.mergeTemplate(ctx, k8sRepo, rai, ra); err != nil {
			return ctrl.Result{}, err
		}

		// apply RAI
		if err := k8sRepo.ApplyReviewAppInstanceFromReviewApp(ctx, rai, ra); err != nil {
			return ctrl.Result{}, err
		}

		// set values for update status
		syncedPullRequests = append(syncedPullRequests, dreamkastv1beta1.ReviewAppStatusSyncedPullRequests{
			Number:                pr.Number,
			ReviewAppInstanceName: rai.Name,
		})
	}

	// delete the ReviewAppInstance that associated PR has already been closed
loop:
	for _, a := range ra.Status.SyncedPullRequests {
		for _, b := range syncedPullRequests {
			if a.Number == b.Number {
				continue loop
			}
		}
		// delete RAI that only exists ResourceStatus
		if err := k8sRepo.DeleteReviewAppInstance(ctx, types.NamespacedName{Name: a.ReviewAppInstanceName, Namespace: ra.Namespace}); err != nil {
			return ctrl.Result{}, err
		}
	}

	// update ReviewApp Status
	ra.Status.SyncedPullRequests = syncedPullRequests
	if err := k8sRepo.UpdateReviewAppStatus(ctx, ra); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (s *ReviewAppService) mergeTemplate(ctx context.Context, k8sRepo repositories.KubernetesRepository, rai *dreamkastv1beta1.ReviewAppInstance, ra *dreamkastv1beta1.ReviewApp) error {
	// construct struct for template's value
	vars := make(map[string]string)
	for i, line := range ra.Spec.Variables {
		idx := strings.Index(line, "=")
		if idx == -1 {
			s.Log.Info(fmt.Sprintf("RA %s: .Spec.Variables[%d] is invalid", ra.Name, i))
			continue
		}
		vars[line[:idx]] = line[idx+1:]
	}
	v := &Value{
		pull_request: pullRequest{
			number: 0, //TODO
		},
		variables: vars,
	}

	// get ApplicationTemplate resource from RA & set to ReviewAppInstance
	at, err := k8sRepo.GetApplicationTemplate(ctx, types.NamespacedName{Name: ra.Spec.Infra.ArgoCDApp.Template, Namespace: ra.Namespace})
	if err != nil {
		return err
	}
	buf, err := yaml.Marshal(&at)
	if err != nil {
		return err
	}
	rai.Spec.Application, err = v.templating(string(buf))
	if err != nil {
		return err
	}

	// get ManifestTemplate resource from RA & set to ReviewAppInstance
	var mts map[string]string
	for _, mtName := range ra.Spec.Infra.Manifests.Templates {
		mt, err := k8sRepo.GetManifestTemplate(ctx, types.NamespacedName{Name: mtName, Namespace: ra.Namespace})
		if err != nil {
			return err
		}
		for key, val := range mt.Data {
			s, err := v.templating(val)
			if err != nil {
				return err
			}
			mts[key] = s
		}
	}

	// set ArgoCDApp.Filepath & Manifests.Dirpath to ReviewAppInstance
	rai.Spec.Infra.ArgoCDApp.Filepath, err = v.templating(ra.Spec.Infra.ArgoCDApp.Filepath)
	if err != nil {
		return err
	}
	rai.Spec.Infra.Manifests.Dirpath, err = v.templating(ra.Spec.Infra.Manifests.Dirpath)
	if err != nil {
		return err
	}

	return nil
}
