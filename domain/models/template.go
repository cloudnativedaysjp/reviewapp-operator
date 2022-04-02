package models

import (
	"bytes"
	"strings"
	"text/template"

	"golang.org/x/xerrors"
)

type Templator struct {
	AppRepo   templateValueAppRepoInfo
	InfraRepo templateValueInfraRepoInfo
	Variables map[string]string
}

type templateValueAppRepoInfo struct {
	Organization     string
	Repository       string
	Branch           string
	PrNumber         int
	LatestCommitHash string
}

type templateValueInfraRepoInfo struct {
	Organization string
	Repository   string
	Branch       string
	// TODO: #56
	//LatestCommitHash string
}

func NewTemplator(
	m ReviewAppOrReviewAppManager,
	pr PullRequest,
) Templator {
	vars := make(map[string]string)
	for _, line := range m.Variables() {
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}
		vars[line[:idx]] = line[idx+1:]
	}
	appTarget := m.AppRepoTarget()
	infraTarget := m.InfraRepoTarget()
	return Templator{
		templateValueAppRepoInfo{
			Organization:     appTarget.Organization,
			Repository:       appTarget.Repository,
			Branch:           pr.Branch,
			PrNumber:         pr.Number,
			LatestCommitHash: pr.LatestCommitHash,
		},
		templateValueInfraRepoInfo{
			Organization: infraTarget.Organization,
			Repository:   infraTarget.Repository,
			Branch:       infraTarget.Branch,
		},
		vars,
	}
}

func (v Templator) WithAppRepoLatestCommitHash(sha string) *Templator {
	v.AppRepo.LatestCommitHash = sha
	return &v
}

func (v Templator) Templating(text string) (string, error) {
	tmpl, err := template.New("Templating").Parse(text)
	if err != nil {
		return "", xerrors.Errorf("Error to parse template: %w", err)
	}
	val := bytes.Buffer{}
	if err := tmpl.Execute(&val, v); err != nil {
		return "", xerrors.Errorf("Error to parse template: %w", err)
	}
	return val.String(), nil
}
