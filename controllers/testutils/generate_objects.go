package testutils

import (
	"os"
	"path/filepath"

	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/yaml"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

const generateObjectsBasePath = "testutils/.objects/"

func GenerateObjects(dirname string) (
	ra dreamkastv1alpha1.ReviewApp,
	pr dreamkastv1alpha1.PullRequest,
	at dreamkastv1alpha1.ApplicationTemplate,
	mt dreamkastv1alpha1.ManifestsTemplate,
	app dreamkastv1alpha1.Application,
	manifests dreamkastv1alpha1.Manifests,
	preStopJt dreamkastv1alpha1.JobTemplate,
	preStopJob batchv1.Job,
) {
	{ // ReviewApp
		raFilePath := filepath.Join(generateObjectsBasePath, dirname, "ra.yaml")
		raBytes, err := os.ReadFile(raFilePath)
		if err == nil {
			_ = yaml.Unmarshal(raBytes, &ra)
		}
	}
	{ // PullRequest
		prFilePath := filepath.Join(generateObjectsBasePath, dirname, "pr.yaml")
		prBytes, err := os.ReadFile(prFilePath)
		if err == nil {
			_ = yaml.Unmarshal(prBytes, &pr)
		}
	}
	{ // ApplicationTemplate
		atFilePath := filepath.Join(generateObjectsBasePath, dirname, "at.yaml")
		atBytes, err := os.ReadFile(atFilePath)
		if err == nil {
			_ = yaml.Unmarshal(atBytes, &at)
		}
	}
	{ // ManifestsTemplate
		mtFilePath := filepath.Join(generateObjectsBasePath, dirname, "mt.yaml")
		mtBytes, err := os.ReadFile(mtFilePath)
		if err == nil {
			_ = yaml.Unmarshal(mtBytes, &mt)
		}
	}
	{ // Application
		appFilePath := filepath.Join(generateObjectsBasePath, dirname, "app.yaml")
		appBytes, err := os.ReadFile(appFilePath)
		if err == nil {
			app = dreamkastv1alpha1.Application(appBytes)
		}
	}
	{ // some manifests from ManifestsTemplate
		manifests = make(map[string]string)
		manifestsDirPath := filepath.Join(generateObjectsBasePath, dirname, "manifests")
		files, _ := os.ReadDir(manifestsDirPath)
		for _, f := range files {
			manifestBytes, err := os.ReadFile(filepath.Join(manifestsDirPath, f.Name()))
			if err == nil {
				manifests[f.Name()] = string(manifestBytes)
			}
		}
	}
	{ // JobTemplate (preStopJob)
		preStopJtFilePath := filepath.Join(generateObjectsBasePath, dirname, "preStopJt.yaml")
		preStopJtBytes, err := os.ReadFile(preStopJtFilePath)
		if err == nil {
			_ = yaml.Unmarshal(preStopJtBytes, &preStopJt)
		}
	}
	{ // Job (preStopJob)
		preStopJobFilePath := filepath.Join(generateObjectsBasePath, dirname, "preStopJob.yaml")
		preStopJobBytes, err := os.ReadFile(preStopJobFilePath)
		if err == nil {
			_ = yaml.Unmarshal(preStopJobBytes, &preStopJob)
		}
	}
	return
}
