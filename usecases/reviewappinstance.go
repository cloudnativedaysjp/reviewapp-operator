package usecases

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ReviewAppInstanceService struct {
	// TODO
}

func (s *ReviewAppInstanceService) ApplyManifestsToInfraRepo(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	////// fetch ReviewApp.infra.repository
	////// create Application manifest from ApplicationTemplate
	////// create manifests from ManifestsTemplate
	////// push

	////// set appRepoPrNum, applicationName, appRepoSha, infraRepoSha to ReviewAppInstance.Status

	return ctrl.Result{}, nil
}

func (s *ReviewAppInstanceService) DeleteManifestsFromInfraRepo(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	////// fetch ReviewApp.infra.repository
	////// delete Application manifest
	////// delete manifests
	////// push

	////// delete element from ReviewAppInstance.Status

	return ctrl.Result{}, nil
}

func (s *ReviewAppInstanceService) NotificationToAppRepoPR(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// for artifact := range ReviewAppStatus.syncedArtifacts

	//// status := check ArgoCD Applications Synced Status
	//// if status.Sha == ReviewAppStatus.syncedArtifacts[].infraRepoSha && !ReviewAppStatus.syncedArtifacts[].notified

	////// notify to PR in app repo
	////// set ReviewAppStatus.syncedArtifacts[].notified = true

	//// endif

	// endfor

	return ctrl.Result{}, nil
}
