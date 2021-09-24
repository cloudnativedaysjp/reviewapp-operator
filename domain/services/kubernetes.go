package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
)

type KubernetesService struct {
	ReviewAppConfigIFace  repositories.ReviewAppConfigIFace
	ReviewAppIFace        repositories.ReviewAppIFace
	ArgoCDApplictionIFace repositories.ArgoCDApplictionIFace
	SecretIFace           repositories.SecretIFace

	Log logr.Logger
}

func NewKubernetes(rac repositories.ReviewAppConfigIFace, ra repositories.ReviewAppIFace, app repositories.ArgoCDApplictionIFace, secret repositories.SecretIFace, logger logr.Logger) *KubernetesService {
	return &KubernetesService{rac, ra, app, secret, logger}
}

func (s *KubernetesService) MergeTemplate(
	ctx context.Context,
	ra *dreamkastv1beta1.ReviewApp,
	ram *dreamkastv1beta1.ReviewAppManager,
	pr *models.PullRequest,
) (err error) {
	// Initialize ReviewApp Spec from ReviewAppManager
	ra.Spec.App = ram.Spec.App
	ra.Spec.Infra = ram.Spec.Infra

	// construct "TemplateValue" model
	vars := make(map[string]string)
	for i, line := range ram.Spec.Variables {
		idx := strings.Index(line, "=")
		if idx == -1 {
			s.Log.Info(fmt.Sprintf("RA %s: .Spec.Variables[%d] is invalid", ram.Name, i))
			continue
		}
		vars[line[:idx]] = line[idx+1:]
	}
	v := models.NewTemplateValue(pr.Organization, pr.Repository, pr.Number, vars)

	// set AppPrNum to ReviewApp
	ra.Spec.AppPrNum = pr.Number

	// set ArgoCDApp.Filepath & Manifests.Dirpath to ReviewApp
	ra.Spec.Infra.ArgoCDApp.Filepath, err = v.Templating(ram.Spec.Infra.ArgoCDApp.Filepath)
	if err != nil {
		return err
	}
	ra.Spec.Infra.Manifests.Dirpath, err = v.Templating(ram.Spec.Infra.Manifests.Dirpath)
	if err != nil {
		return err
	}

	// get ApplicationTemplate & ManifestTemplate resource from RA & set to ReviewApp
	rac, err := s.ReviewAppConfigIFace.GetReviewAppConfig(ctx, ram.Namespace, ram.Name)
	if err != nil {
		return err
	}
	ra.Spec.Application, err = v.Templating(rac.ApplicationTemplate.Spec.Template)
	if err != nil {
		return err
	}
	mts := make(map[string]string)
	for key, val := range rac.ManifestsTemplate {
		s, err := v.Templating(val)
		if err != nil {
			return err
		}
		mts[key] = s
	}
	ra.Spec.Manifests = mts

	ra.Spec.App.Message, err = v.Templating(ram.Spec.App.Message)
	if err != nil {
		return err
	}

	return nil
}

func (s *KubernetesService) ApplyReviewAppFromReviewAppManager(
	ctx context.Context,
	ra *dreamkastv1beta1.ReviewApp,
	ram *dreamkastv1beta1.ReviewAppManager,
) error {
	raModel := models.ReviewApp{ReviewApp: *ra}
	return s.ReviewAppIFace.ApplyReviewAppWithOwnerRef(ctx, &raModel, ram)
}

func (s *KubernetesService) GetReviewAppManagerFromReviewApp(
	ctx context.Context,
	ra *dreamkastv1beta1.ReviewApp,
) (*dreamkastv1beta1.ReviewAppManager, error) {
	raModel := models.ReviewApp{ReviewApp: *ra}
	ramModel, err := s.ReviewAppIFace.GetReviewAppManagerFromReviewApp(ctx, &raModel)
	if err != nil {
		return nil, err
	}
	return &ramModel.ReviewAppManager, nil
}

func (s *KubernetesService) UpdateReviewAppManagerStatus(
	ctx context.Context,
	ram *dreamkastv1beta1.ReviewAppManager,
) error {
	rac, err := s.ReviewAppConfigIFace.GetReviewAppConfig(ctx, ram.Namespace, ram.Name)
	if err != nil {
		return err
	}
	rac.ReviewAppManager.Status = ram.Status
	return s.ReviewAppConfigIFace.UpdateReviewAppManagerStatus(ctx, rac)
}

func (s *KubernetesService) UpdateReviewAppStatus(
	ctx context.Context,
	ra *dreamkastv1beta1.ReviewApp,
) error {
	return s.ReviewAppIFace.UpdateReviewAppStatus(ctx, &models.ReviewApp{ReviewApp: *ra})
}

func (s *KubernetesService) GetArgoCDApplicationWithAnnotations(ctx context.Context, applicationStr, organization, repository, commitHashInAppRepo string) (string, error) {
	var err error

	// set annotations
	appWithAnnotation := models.ArgoCDApplicationString(applicationStr)
	appWithAnnotation, err = s.ArgoCDApplictionIFace.PrintArgoCDApplicationWithAnnotation(
		ctx, appWithAnnotation,
		models.AnnotationAppOrgNameForArgoCDApplication, organization,
	)
	if err != nil {
		return "", err
	}
	appWithAnnotation, err = s.ArgoCDApplictionIFace.PrintArgoCDApplicationWithAnnotation(
		ctx, appWithAnnotation,
		models.AnnotationAppRepoNameForArgoCDApplication, repository,
	)
	if err != nil {
		return "", err
	}
	appWithAnnotation, err = s.ArgoCDApplictionIFace.PrintArgoCDApplicationWithAnnotation(
		ctx, appWithAnnotation,
		models.AnnotationAppCommitHashForArgoCDApplication, commitHashInAppRepo,
	)
	if err != nil {
		return "", err
	}

	return string(appWithAnnotation), nil
}

func (s *KubernetesService) GetArgoCDApplicationName(ctx context.Context, applicationStr string) (string, error) {
	name, err := s.ArgoCDApplictionIFace.PrintArgoCDApplicationName(ctx, models.ArgoCDApplicationString(applicationStr))
	if err != nil {
		return "", err
	}
	return name, nil
}

func (s *KubernetesService) GetArgoCDApplicationNamespace(ctx context.Context, applicationStr string) (string, error) {
	name, err := s.ArgoCDApplictionIFace.PrintArgoCDApplicationNamespace(ctx, models.ArgoCDApplicationString(applicationStr))
	if err != nil {
		return "", err
	}
	return name, nil
}

func (s *KubernetesService) AddFinalizersToReviewApp(ctx context.Context, ra *dreamkastv1beta1.ReviewApp, finalizers ...string) error {
	return s.ReviewAppConfigIFace.AddFinalizersToReviewApp(ctx, &models.ReviewApp{ReviewApp: *ra}, finalizers...)
}

func (s *KubernetesService) RemoveFinalizersToReviewApp(ctx context.Context, ra *dreamkastv1beta1.ReviewApp, finalizers ...string) error {
	return s.ReviewAppConfigIFace.RemoveFinalizersToReviewApp(ctx, &models.ReviewApp{ReviewApp: *ra}, finalizers...)
}
