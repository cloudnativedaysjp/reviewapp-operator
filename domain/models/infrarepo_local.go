package models

import "fmt"

type InfraRepoLocalDir struct {
	baseDir          string
	latestCommitHash string
}

func NewInfraRepoLocal(baseDir string) InfraRepoLocalDir {
	return InfraRepoLocalDir{baseDir: baseDir}
}

func (m InfraRepoLocalDir) SetLatestCommitHash(latestCommitHash string) InfraRepoLocalDir {
	m.latestCommitHash = latestCommitHash
	return m
}

func (m InfraRepoLocalDir) BaseDir() string {
	return m.baseDir
}

func (m InfraRepoLocalDir) LatestCommitHash() string {
	return m.latestCommitHash
}

func (m InfraRepoLocalDir) CommitMsgUpdate(ra ReviewApp) string {
	return fmt.Sprintf(
		"Automatic update by cloudnativedays/reviewapp-operator (%s/%s@%s)",
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Status.Sync.SyncedPullRequest.LatestCommitHash,
	)
}

func (m InfraRepoLocalDir) CommitMsgDeletion(ra ReviewApp) string {
	return fmt.Sprintf(
		"Automatic GC by cloudnativedays/reviewapp-operator (%s/%s@%s)",
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Status.Sync.SyncedPullRequest.LatestCommitHash,
	)
}
