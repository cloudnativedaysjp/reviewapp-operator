package template

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
	Branch          string
	PrNumber        int
	LatestCommitSha string
}

type TemplateValueInfraRepoInfo struct {
	Organization string
	Repository   string
	// TODO: #56
	//LatestCommitSha string
}

func NewTemplateValue(
	appOrg, appRepo, appBranch string, appPrNum int,
	infraOrg, infraRepo string,
	variables map[string]string,
) *TemplateValue {
	return &TemplateValue{
		TemplateValueAppRepoInfo{
			Organization: appOrg,
			Repository:   appRepo,
			Branch:       appBranch,
			PrNumber:     appPrNum,
		},
		TemplateValueInfraRepoInfo{
			Organization: infraOrg,
			Repository:   infraRepo,
		},
		variables,
	}
}

func (v TemplateValue) WithAppRepoLatestCommitSha(sha string) *TemplateValue {
	v.AppRepo.LatestCommitSha = sha
	return &v
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

func (v TemplateValue) MapTemplating(m map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for key, val := range m {
		s, err := v.Templating(val)
		if err != nil {
			return nil, err
		}
		result[key] = s
	}
	return result, nil
}

func (v TemplateValue) MapTemplatingAndAppend(base, m map[string]string) (map[string]string, error) {
	appended, err := v.MapTemplating(m)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for k, v := range base {
		result[k] = v
	}
	for k, v := range appended {
		result[k] = v
	}
	return result, nil
}
