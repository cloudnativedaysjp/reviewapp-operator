package repositories

import "github.com/go-logr/logr"

type GitApiFactory interface {
	NewRepository(string, string, logr.Logger) (GitApiRepository, error)
}

type GitApiRepository interface {
	// TODO
}
