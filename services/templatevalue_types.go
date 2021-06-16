package services

import (
	"bytes"
	"text/template"
)

var (
	t = template.New("services")
)

type Value struct {
	pull_request pullRequest
	variables    map[string]string
}

func (v Value) templating(text string) (string, error) {
	tmpl, err := t.Parse(text)
	if err != nil {
		return "", err
	}
	val := new(bytes.Buffer)
	if err := tmpl.Execute(val, v); err != nil {
		return "", err
	}
	return val.String(), nil
}

type pullRequest struct {
	number uint
}
