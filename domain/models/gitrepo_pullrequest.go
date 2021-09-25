package models

const CandidateLabelName = "candidate-template"

type PullRequest struct {
	Organization  string
	Repository    string
	Number        int
	HeadCommitSha string
	Labels        []string // TODO
}

func NewPullRequest(organization, repository string, number int, headCommitSha string, labels []string) *PullRequest {
	return &PullRequest{
		Organization:  organization,
		Repository:    repository,
		Number:        number,
		HeadCommitSha: headCommitSha,
		Labels:        labels,
	}
}
