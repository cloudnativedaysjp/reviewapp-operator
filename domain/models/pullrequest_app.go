package models

type PullRequestApp struct {
	pullRequestCommon
}

func NewPullRequestApp(organization, repository string, number int, headCommitSha string) *PullRequestApp {
	return &PullRequestApp{
		pullRequestCommon: pullRequestCommon{
			Organization:  organization,
			Repository:    repository,
			Number:        number,
			HeadCommitSha: headCommitSha,
		},
	}
}
