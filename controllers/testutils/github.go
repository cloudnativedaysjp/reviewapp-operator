package testutils

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var ( // 参照渡しするため var で定義
	prStateOpen   string = "open"
	prStateClosed string = "closed"
)

type GitHubClient struct {
	ts oauth2.TokenSource
}

func NewGitHubClient(token string) *GitHubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return &GitHubClient{ts}
}

func (c GitHubClient) ClosePr(org, repo string, prNum int) error {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, c.ts))
	pr, _, err := client.PullRequests.Get(ctx, org, repo, prNum)
	if err != nil {
		return err
	}
	if *pr.State == prStateOpen {
		if _, _, err := client.PullRequests.Edit(ctx, org, repo, prNum, &github.PullRequest{State: &prStateClosed}); err != nil {
			return err
		}
	}
	return nil
}

func (c GitHubClient) OpenPr(org, repo string, prNum int) error {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, c.ts))
	pr, _, err := client.PullRequests.Get(ctx, org, repo, prNum)
	if err != nil {
		return err
	}
	if *pr.State == prStateClosed {
		if _, _, err := client.PullRequests.Edit(ctx, org, repo, prNum, &github.PullRequest{State: &prStateOpen}); err != nil {
			return err
		}
	}
	return nil
}

func (c GitHubClient) GetLatestMessage(org, repo string, prNum int) (string, error) {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, c.ts))
	comments, _, err := client.Issues.ListComments(ctx, org, repo, prNum, &github.IssueListCommentsOptions{
		// Sort: "created", Direction: "desc",  // NOTE: これらのオプションが効かないので全件取得後に最新1件を取り出している。
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 1048576,
		},
	})
	if err != nil {
		return "", err
	} else if len(comments) == 0 {
		return "", fmt.Errorf("PR is not commented")
	}
	return *comments[len(comments)-1].Body, nil
}

func (c GitHubClient) GetUpdatedFilenamesInLatestCommit(org, repo, branch string) ([]string, error) {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, c.ts))
	ref, _, err := client.Git.GetRef(ctx, org, repo, fmt.Sprintf("heads/%s", branch))
	if err != nil {
		return nil, err
	}
	commit, _, err := client.Repositories.GetCommit(ctx, org, repo, *ref.Object.SHA)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, file := range commit.Files {
		if *file.Status == "added" || *file.Status == "modified" {
			result = append(result, *file.Filename)
		}
	}
	return result, nil
}

func (c GitHubClient) GetDeletedFilenamesInLatestCommit(org, repo, branch string) ([]string, error) {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, c.ts))
	ref, _, err := client.Git.GetRef(ctx, org, repo, fmt.Sprintf("heads/%s", branch))
	if err != nil {
		return nil, err
	}
	commit, _, err := client.Repositories.GetCommit(ctx, org, repo, *ref.Object.SHA)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, file := range commit.Files {
		// TODO: debug
		fmt.Println(*file.Status)
		if *file.Status == "deleted" {
			result = append(result, *file.Filename)
		}
	}
	return result, nil
}
