package usecases

import (
	"bytes"
	"text/template"
)

var (
	t = template.New("services")
)

type Value struct {
	PullRequest PullRequest
	Variables   map[string]string
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

type PullRequest struct {
	Number int
}
