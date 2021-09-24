package models

import (
	"bytes"
	"text/template"
)

var (
	t = template.New("services")
)

type TemplateValue struct {
	PullRequest TemplatePullRequest
	Variables   map[string]string
}

type TemplatePullRequest struct {
	Organization string
	Repository   string
	Number       int
}

func NewTemplateValue(organization, repository string, prNum int, variables map[string]string) *TemplateValue {
	return &TemplateValue{TemplatePullRequest{organization, repository, prNum}, variables}
}

func (v TemplateValue) Templating(text string) (string, error) {
	tmpl, err := template.New("Templating").Parse(text)
	if err != nil {
		return "", err
	}
	val := bytes.Buffer{}
	if err := tmpl.Execute(&val, v); err != nil {
		return "", err
	}
	return val.String(), nil
}
