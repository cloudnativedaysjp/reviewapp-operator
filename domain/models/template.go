package models

import (
	"bytes"
	"text/template"

	"golang.org/x/xerrors"
)

type TemplateValue struct {
	AppRepo   TemplateValueAppRepoInfo
	InfraRepo TemplateValueInfraRepoInfo
	Variables map[string]string
}

type TemplateValueAppRepoInfo struct {
	Organization    string
	Repository      string
	PrNumber        int
	LatestCommitSha string
}

type TemplateValueInfraRepoInfo struct {
	Organization    string
	Repository      string
	LatestCommitSha string
}

func NewTemplateValue(
	appOrg, appRepo string, appPrNum int, appLatestCommitSha string,
	infraOrg, infraRepo, infraLatestCommitSha string,
	variables map[string]string,
) *TemplateValue {
	return &TemplateValue{
		TemplateValueAppRepoInfo{appOrg, appRepo, appPrNum, appLatestCommitSha},
		TemplateValueInfraRepoInfo{infraOrg, infraRepo, infraLatestCommitSha},
		variables,
	}
}

func (v TemplateValue) Templating(text string) (string, error) {
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
