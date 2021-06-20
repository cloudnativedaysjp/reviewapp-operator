package repositories

import "io"

type PullRequestInfra struct {
	Iface PullRequestInfraIFace
}

type PullRequestInfraIFace interface {
	WithCredential(string) error
	CheckDirectoryExistence(org, repo string, dirname string) error
	WithCreateFile(org, repo string, filename string, contents io.Reader) error
	WithDeleteFile(org, repo string, filename string) error
	Commit(message string) error
}

func NewPullRequestInfra(iface PullRequestInfraIFace) *PullRequestInfra {
	return &PullRequestInfra{iface}
}
