package models

type PullRequest struct {
	Organization  string
	Repository    string
	Number        int
	HeadCommitSha string
}

func NewPullRequest(organization, repository string, number int, headCommitSha string) *PullRequest {
	return &PullRequest{
		Organization:  organization,
		Repository:    repository,
		Number:        number,
		HeadCommitSha: headCommitSha,
	}
}
