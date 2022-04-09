package models

import "regexp"

const (
	candidateLabelName = "candidate-template"
)

type PullRequest struct {
	Organization  string
	Repository    string
	Branch        string
	Number        int
	HeadCommitHash string
	Title         string
	Labels        []string
}

func NewPullRequest(organization, repository, branch string, number int, headCommitHash string, title string, labels []string) PullRequest {
	return PullRequest{
		Organization:  organization,
		Repository:    repository,
		Branch:        branch,
		Number:        number,
		HeadCommitHash: headCommitHash,
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

type PullRequests []PullRequest

func (m PullRequests) ExcludeSpecificPR(ra ReviewAppOrReviewAppManager) PullRequests {
	ignoreLabels := ra.AppRepoTarget().IgnoreLabels
	ignoreTitleExp := ra.AppRepoTarget().IgnoreTitleExp
	removedCount := 0
	for idx, pr := range m {
		for _, actualLabel := range pr.Labels {
			for _, ignoreLabel := range ignoreLabels {
				if actualLabel == ignoreLabel {
					m = m.remove(idx - removedCount)
					removedCount += 1
				}
			}
		}
	}
	removedCount = 0
	if ignoreTitleExp != "" {
		r := regexp.MustCompile(ignoreTitleExp)
		for idx, pr := range m {
			if r.Match([]byte(pr.Title)) {
				m = m.remove(idx - removedCount)
				removedCount += 1
			}
		}
	}
	return m
}

func (m PullRequests) remove(idx int) PullRequests {
	if idx == len(m)-1 {
		return m[:idx]
	} else {
		return append(m[:idx], m[idx+1:]...)
	}
}
