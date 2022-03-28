package models

import (
	"path/filepath"
)

type File struct {
	Filepath string
	Content  []byte
}

func NewFileFromApplication(ra ReviewApp, application Application, pr PullRequest, l InfraRepoLocalDir) File {
	return File{Filepath: filepath.Join(l.baseDir, ra.GetAtFilepath()), Content: []byte(application)}
}

func NewFilesFromManifests(ra ReviewApp, manifests Manifests, pr PullRequest, l InfraRepoLocalDir) []File {
	var result []File
	for mtFilename, mtContent := range manifests {
		result = append(result, File{Filepath: filepath.Join(l.baseDir, ra.GetMtDirpath(), mtFilename), Content: []byte(mtContent)})
	}
	return result
}

func NewFileFromName(l InfraRepoLocalDir, filename string) File {
	return File{Filepath: filepath.Join(l.baseDir, filename)}
}
