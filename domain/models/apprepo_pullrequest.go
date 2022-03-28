package models

const (
	candidateLabelName = "candidate-template"
)

type PullRequest struct {
	Organization  string
	Repository    string
	Branch        string
	Number        int
	HeadCommitSha string
	Title         string
	Labels        []string
}

func NewPullRequest(organization, repository, branch string, number int, headCommitSha string, title string, labels []string) PullRequest {
	return PullRequest{
		Organization:  organization,
		Repository:    repository,
		Branch:        branch,
		Number:        number,
		HeadCommitSha: headCommitSha,
		Title:         title,
		Labels:        labels,
	}
}

func (m PullRequest) IsCandidate() bool {
	isCandidate := false
	for _, l := range m.Labels {
		if l == candidateLabelName {
			isCandidate = true
		}
	}
	return isCandidate
}
