package models

type pullRequestCommon struct {
	Organization  string
	Repository    string
	Number        int
	HeadCommitSha string
}
