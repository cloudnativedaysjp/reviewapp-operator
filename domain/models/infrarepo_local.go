package models

import "fmt"

type InfraRepoLocalDir struct {
	baseDir         string
	latestCommitSha string
}

func NewInfraRepoLocal(baseDir string) InfraRepoLocalDir {
	return InfraRepoLocalDir{baseDir: baseDir}
}

func (m InfraRepoLocalDir) SetLatestCommitSha(latestCommitSha string) InfraRepoLocalDir {
	m.latestCommitSha = latestCommitSha
	return m
}

func (m InfraRepoLocalDir) GetBaseDir() string {
	return m.baseDir
}

func (m InfraRepoLocalDir) GetLatestCommitSha() string {
	return m.latestCommitSha
}

func (m InfraRepoLocalDir) GetCommitMsgUpdate(ra ReviewApp) string {
	return fmt.Sprintf(
		"Automatic update by cloudnativedays/reviewapp-operator (%s/%s@%s)",
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Status.Sync.AppRepoLatestCommitSha,
	)
}

func (m InfraRepoLocalDir) GetCommitMsgDeletion(ra ReviewApp) string {
	return fmt.Sprintf(
		"Automatic GC by cloudnativedays/reviewapp-operator (%s/%s@%s)",
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Status.Sync.AppRepoLatestCommitSha,
	)
}
