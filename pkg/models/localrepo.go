package models

import (
	"fmt"
	"path/filepath"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

type InfraRepoLocalDir struct {
	baseDir          string
	latestCommitHash string
}

func NewInfraRepoLocalDir(baseDir string) InfraRepoLocalDir {
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

func (m InfraRepoLocalDir) CommitMsgUpdate(ra dreamkastv1alpha1.ReviewApp) string {
	return fmt.Sprintf(
		"Automatic update by cloudnativedays/reviewapp-operator (%s/%s@%s)",
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Status.PullRequestCache.LatestCommitHash,
	)
}

func (m InfraRepoLocalDir) CommitMsgDeletion(ra dreamkastv1alpha1.ReviewApp) string {
	return fmt.Sprintf(
		"Automatic GC by cloudnativedays/reviewapp-operator (%s/%s@%s)",
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Status.PullRequestCache.LatestCommitHash,
	)
}

type File struct {
	BaseDir  string
	Filepath string
	Content  []byte
}

func NewFileFromApplication(
	ra dreamkastv1alpha1.ReviewApp, application dreamkastv1alpha1.Application,
	pr dreamkastv1alpha1.PullRequest, l InfraRepoLocalDir,
) File {
	return File{
		BaseDir:  l.baseDir,
		Filepath: ra.Spec.InfraConfig.ArgoCDApp.Filepath,
		Content:  []byte(application),
	}
}

func NewFilesFromManifests(
	ra dreamkastv1alpha1.ReviewApp, manifests dreamkastv1alpha1.Manifests,
	pr dreamkastv1alpha1.PullRequest, l InfraRepoLocalDir,
) []File {
	var result []File
	for mtFilename, mtContent := range manifests {
		result = append(result, File{
			BaseDir:  l.baseDir,
			Filepath: filepath.Join(ra.Spec.InfraConfig.Manifests.Dirpath, mtFilename),
			Content:  []byte(mtContent),
		})
	}
	return result
}

func NewFileFromName(l InfraRepoLocalDir, filename string) File {
	return File{BaseDir: l.baseDir, Filepath: filepath.Join(l.baseDir, filename)}
}
