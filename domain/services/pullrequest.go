package services

import (
	"context"
	"time"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils"
)

const (
	pullRequestResyncPeriod = time.Minute
)

type PullRequestServiceIface interface {
	Get(context.Context, models.ReviewApp, models.GitCredential, *utils.DatetimeFactory) (models.PullRequest, error)
}

type PullRequestService struct {
	GitApiRepository repositories.GitAPI
}

func NewPullRequestService(gitApi repositories.GitAPI) *PullRequestService {
	return &PullRequestService{gitApi}
}

func (s PullRequestService) Get(ctx context.Context, ra models.ReviewApp, cred models.GitCredential, f *utils.DatetimeFactory) (models.PullRequest, error) {
	appRepoTarget := ra.AppRepoTarget()
	raStatus := ra.GetStatus()
	now := f.Now()
	// check previous synced timestamp
	previousSyncedTimestamp := raStatus.Sync.SyncedPullRequest.SyncTimestamp
	if previousSyncedTimestamp != "" {
		t, err := utils.NewDatetime(previousSyncedTimestamp)
		if err != nil {
			return models.PullRequest{}, err
		}
		// if dont need resync, return values from ReviewApp Object
		if !t.Before(now, pullRequestResyncPeriod) {
			return models.NewPullRequest(
				appRepoTarget.Organization, appRepoTarget.Repository, raStatus.Sync.SyncedPullRequest.Branch,
				ra.PrNum(), raStatus.Sync.SyncedPullRequest.LatestCommitHash,
				raStatus.Sync.SyncedPullRequest.Title, raStatus.Sync.SyncedPullRequest.Labels,
			), nil
		}
	}
	// otherwise, get from GitAPI repository & update timestamp
	if err := s.GitApiRepository.WithCredential(cred); err != nil {
		return models.PullRequest{}, err
	}
	pr, err := s.GitApiRepository.GetPullRequest(ctx, appRepoTarget, ra.PrNum())
	if err != nil {
		return models.PullRequest{}, err
	}
	raStatus.Sync.SyncedPullRequest.SyncTimestamp = now.ToString()
	return pr, nil
}
