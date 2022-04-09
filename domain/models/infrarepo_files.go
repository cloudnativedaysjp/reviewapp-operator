package models

import (
	"path/filepath"
)

type File struct {
	BaseDir  string
	Filepath string
	Content  []byte
}

func NewFileFromApplication(ra ReviewApp, application Application, pr PullRequest, l InfraRepoLocalDir) File {
	return File{BaseDir: l.baseDir, Filepath: ra.AtFilepath(), Content: []byte(application)}
}

func NewFilesFromManifests(ra ReviewApp, manifests Manifests, pr PullRequest, l InfraRepoLocalDir) []File {
	var result []File
	for mtFilename, mtContent := range manifests {
		result = append(result, File{BaseDir: l.baseDir, Filepath: filepath.Join(ra.MtDirpath(), mtFilename), Content: []byte(mtContent)})
	}
	return result
}

func NewFileFromName(l InfraRepoLocalDir, filename string) File {
	return File{BaseDir: l.baseDir, Filepath: filepath.Join(l.baseDir, filename)}
}
