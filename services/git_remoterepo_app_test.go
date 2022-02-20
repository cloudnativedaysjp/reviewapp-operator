package services

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	"github.com/cloudnativedaysjp/reviewapp-operator/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func TestGitRemoteRepoAppService_ListOpenPullRequestWithSpecificConditions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testUsername := "testuser"
	testToken := "testtoken"
	testOrg := "testorg"
	testRepo := "testrepo"
	testPrNum := 1
	testBranchName := "testbranch"
	testHeadCommitSha := "1234567"
	testPrTitle := "title-ok"
	testPrTitleIgnored := "title-ignore"
	testPrLabel := "label-ok"
	testPrLabelIgnored := "label-ignore"

	type fields struct {
		gitapi func() gateways.GitHubIFace
	}
	type args struct {
		ctx            context.Context
		org            string
		repo           string
		username       string
		token          string
		ignoreLabels   []string
		ignoreTitleExp string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*gateways.PullRequest
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ListOpenPullRequests(context.Background(), testOrg, testRepo).Return(
					[]*gateways.PullRequest{{
						Organization:  testOrg,
						Repository:    testRepo,
						Number:        testPrNum,
						Branch:        testBranchName,
						HeadCommitSha: testHeadCommitSha,
						Title:         testPrTitle,
					}}, nil,
				)
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{},
				ignoreTitleExp: "",
			},
			want: []*gateways.PullRequest{{
				Organization:  testOrg,
				Repository:    testRepo,
				Number:        testPrNum,
				Branch:        testBranchName,
				HeadCommitSha: testHeadCommitSha,
				Title:         testPrTitle,
			}},
			wantErr: false,
		},
		{
			name: "normal(ignoreLabels)_01",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ListOpenPullRequests(context.Background(), testOrg, testRepo).Return(
					[]*gateways.PullRequest{
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitle,
							Labels:        []string{testPrLabel},
						},
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitle,
							Labels:        []string{testPrLabelIgnored},
						},
					}, nil,
				)
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{testPrLabelIgnored},
				ignoreTitleExp: "",
			},
			want: []*gateways.PullRequest{{
				Organization:  testOrg,
				Repository:    testRepo,
				Number:        testPrNum,
				Branch:        testBranchName,
				HeadCommitSha: testHeadCommitSha,
				Title:         testPrTitle,
				Labels:        []string{testPrLabel},
			}},
			wantErr: false,
		},
		{
			name: "normal(ignoreLabels)_02",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ListOpenPullRequests(context.Background(), testOrg, testRepo).Return(
					[]*gateways.PullRequest{{
						Organization:  testOrg,
						Repository:    testRepo,
						Number:        testPrNum,
						Branch:        testBranchName,
						HeadCommitSha: testHeadCommitSha,
						Title:         testPrTitle,
						Labels:        []string{testPrLabelIgnored},
					}}, nil,
				)
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{testPrLabelIgnored},
				ignoreTitleExp: "",
			},
			want:    []*gateways.PullRequest{},
			wantErr: false,
		},
		{
			name: "normal(ignoreLabels)_03_multi-ignored-pr",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ListOpenPullRequests(context.Background(), testOrg, testRepo).Return(
					[]*gateways.PullRequest{
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitle,
							Labels:        []string{testPrLabelIgnored},
						},
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitle,
							Labels:        []string{testPrLabelIgnored},
						},
					}, nil,
				)
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{testPrLabelIgnored},
				ignoreTitleExp: "",
			},
			want:    []*gateways.PullRequest{},
			wantErr: false,
		},
		{
			name: "normal(ignoreTitleExp)_01",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ListOpenPullRequests(context.Background(), testOrg, testRepo).Return(
					[]*gateways.PullRequest{
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitle,
						},
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitleIgnored,
						},
					}, nil,
				)
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{},
				ignoreTitleExp: testPrTitleIgnored,
			},
			want: []*gateways.PullRequest{{
				Organization:  testOrg,
				Repository:    testRepo,
				Number:        testPrNum,
				Branch:        testBranchName,
				HeadCommitSha: testHeadCommitSha,
				Title:         testPrTitle,
			}},
			wantErr: false,
		},
		{
			name: "normal(ignoreTitleExp)_02",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ListOpenPullRequests(context.Background(), testOrg, testRepo).Return(
					[]*gateways.PullRequest{{
						Organization:  testOrg,
						Repository:    testRepo,
						Number:        testPrNum,
						Branch:        testBranchName,
						HeadCommitSha: testHeadCommitSha,
						Title:         testPrTitleIgnored,
					}}, nil,
				)
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{},
				ignoreTitleExp: testPrTitleIgnored,
			},
			want:    []*gateways.PullRequest{},
			wantErr: false,
		},
		{
			name: "normal(ignoreTitleExp)_03_multi-ignored-pr",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ListOpenPullRequests(context.Background(), testOrg, testRepo).Return(
					[]*gateways.PullRequest{
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitleIgnored,
						},
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitleIgnored,
						},
					}, nil,
				)
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{},
				ignoreTitleExp: testPrTitleIgnored,
			},
			want:    []*gateways.PullRequest{},
			wantErr: false,
		},
		{
			name: "normal(ignoreLabel&ignoreTitleExp)_01",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ListOpenPullRequests(context.Background(), testOrg, testRepo).Return(
					[]*gateways.PullRequest{
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitle,
							Labels:        []string{testPrLabelIgnored},
						},
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitleIgnored,
						},
					}, nil,
				)
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{testPrLabelIgnored},
				ignoreTitleExp: testPrTitleIgnored,
			},
			want:    []*gateways.PullRequest{},
			wantErr: false,
		},
		{
			name: "normal(ignoreLabel&ignoreTitleExp)_01",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ListOpenPullRequests(context.Background(), testOrg, testRepo).Return(
					[]*gateways.PullRequest{
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitle,
							Labels:        []string{testPrLabelIgnored},
						},
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitleIgnored,
						},
					}, nil,
				)
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{testPrLabelIgnored},
				ignoreTitleExp: testPrTitleIgnored,
			},
			want:    []*gateways.PullRequest{},
			wantErr: false,
		},
		{
			name: "normal(ignoreLabel&ignoreTitleExp)_02",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ListOpenPullRequests(context.Background(), testOrg, testRepo).Return(
					[]*gateways.PullRequest{
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitleIgnored,
						},
						{
							Organization:  testOrg,
							Repository:    testRepo,
							Number:        testPrNum,
							Branch:        testBranchName,
							HeadCommitSha: testHeadCommitSha,
							Title:         testPrTitle,
							Labels:        []string{testPrLabelIgnored},
						},
					}, nil,
				)
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{testPrLabelIgnored},
				ignoreTitleExp: testPrTitleIgnored,
			},
			want:    []*gateways.PullRequest{},
			wantErr: false,
		},
		{
			name: "abnormal(invalid_credential)",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(fmt.Errorf("invalid credential"))
				return c
			}},
			args: args{
				ctx:            context.Background(),
				org:            testOrg,
				repo:           testRepo,
				username:       testUsername,
				token:          testToken,
				ignoreLabels:   []string{},
				ignoreTitleExp: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GitRemoteRepoAppService{
				gitapi: tt.fields.gitapi(),
			}
			got, err := s.ListOpenPullRequestWithSpecificConditions(tt.args.ctx, tt.args.org, tt.args.repo, tt.args.username, tt.args.token, tt.args.ignoreLabels, tt.args.ignoreTitleExp)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitRemoteRepoAppService.ListOpenPullRequestWithSpecificConditions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GitRemoteRepoAppService.ListOpenPullRequestWithSpecificConditions(): %v", diff)
			}
		})
	}
}

func TestGitRemoteRepoAppService_GetPullRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testUsername := "testuser"
	testToken := "testtoken"
	testOrg := "testorg"
	testRepo := "testrepo"
	testPrNum := 1
	testBranchName := "testbranch"
	testHeadCommitSha := "1234567"

	type fields struct {
		gitapi func() gateways.GitHubIFace
	}
	type args struct {
		ctx      context.Context
		org      string
		repo     string
		prNum    int
		username string
		token    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *gateways.PullRequest
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().GetPullRequest(context.Background(), testOrg, testRepo, testPrNum).Return(
					&gateways.PullRequest{
						Organization:  testOrg,
						Repository:    testRepo,
						Number:        testPrNum,
						Branch:        testBranchName,
						HeadCommitSha: testHeadCommitSha,
					}, nil,
				)
				return c
			}},
			args: args{
				ctx:      context.Background(),
				org:      testOrg,
				repo:     testRepo,
				prNum:    testPrNum,
				username: testUsername,
				token:    testToken,
			},
			want: &gateways.PullRequest{
				Organization:  testOrg,
				Repository:    testRepo,
				Number:        testPrNum,
				Branch:        testBranchName,
				HeadCommitSha: testHeadCommitSha,
			},
			wantErr: false,
		},
		{
			name: "abnormal (invalid credential)",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(fmt.Errorf("invalid credential"))
				return c
			}},
			args: args{
				ctx:      context.Background(),
				org:      testOrg,
				repo:     testRepo,
				prNum:    testPrNum,
				username: testUsername,
				token:    testToken,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GitRemoteRepoAppService{
				gitapi: tt.fields.gitapi(),
			}
			got, err := s.GetPullRequest(tt.args.ctx, tt.args.org, tt.args.repo, tt.args.prNum, tt.args.username, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitRemoteRepoAppService.GetPullRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GitRemoteRepoAppService.GetPullRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitRemoteRepoAppService_IsCandidatePr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		gitapi func() gateways.GitHubIFace
	}
	type args struct {
		pr *gateways.PullRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "target label contains",
			fields: fields{func() gateways.GitHubIFace {
				return mock.NewMockGitHubIFace(ctrl)
			}},
			args: args{
				pr: &gateways.PullRequest{Labels: []string{"hoge", "fuga", "piyo", gateways.CandidateLabelName}},
			},
			want: true,
		},
		{
			name: "target label does not contain",
			fields: fields{func() gateways.GitHubIFace {
				return mock.NewMockGitHubIFace(ctrl)
			}},
			args: args{
				pr: &gateways.PullRequest{Labels: []string{"hoge", "fuga", "piyo"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GitRemoteRepoAppService{
				gitapi: tt.fields.gitapi(),
			}
			if got := s.IsCandidatePr(tt.args.pr); got != tt.want {
				t.Errorf("GitRemoteRepoAppService.IsCandidatePr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitRemoteRepoAppService_SendMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testUsername := "testuser"
	testToken := "testtoken"
	testOrg := "testorg"
	testRepo := "testrepo"
	testPrNum := 1
	testBranchName := "testbranch"
	testHeadCommitSha := "1234567"
	testMessage := "testmessage"

	type fields struct {
		gitapi func() gateways.GitHubIFace
	}
	type args struct {
		ctx      context.Context
		pr       *gateways.PullRequest
		message  string
		username string
		token    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().CommentToPullRequest(context.Background(), gateways.PullRequest{
					Organization:  testOrg,
					Repository:    testRepo,
					Number:        testPrNum,
					Branch:        testBranchName,
					HeadCommitSha: testHeadCommitSha,
				}, testMessage).Return(nil)
				return c
			}},
			args: args{
				ctx: context.Background(),
				pr: &gateways.PullRequest{
					Organization:  testOrg,
					Repository:    testRepo,
					Number:        testPrNum,
					Branch:        testBranchName,
					HeadCommitSha: testHeadCommitSha,
				},
				message:  testMessage,
				username: testUsername,
				token:    testToken,
			},
			wantErr: false,
		},
		{
			name: "abnormal (invalid credential)",
			fields: fields{func() gateways.GitHubIFace {
				c := mock.NewMockGitHubIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(fmt.Errorf("invalid credential"))
				return c
			}},
			args: args{
				ctx: context.Background(),
				pr: &gateways.PullRequest{
					Organization:  testOrg,
					Repository:    testRepo,
					Number:        testPrNum,
					Branch:        testBranchName,
					HeadCommitSha: testHeadCommitSha,
				},
				message:  testMessage,
				username: testUsername,
				token:    testToken,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GitRemoteRepoAppService{
				gitapi: tt.fields.gitapi(),
			}
			if err := s.SendMessage(tt.args.ctx, tt.args.pr, tt.args.message, tt.args.username, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("GitRemoteRepoAppService.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
