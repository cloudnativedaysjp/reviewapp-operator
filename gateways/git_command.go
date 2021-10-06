package gateways

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudnativedaysjp/reviewapp-operator/models"
	"github.com/go-logr/logr"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
	"k8s.io/utils/exec"
)

const (
	baseURL      = `https://%s:%s@github.com`
	noreplyEmail = `%s@users.noreply.github.com`
)

type GitCommandIFace interface {
	WithCredential(username, token string) error
	Pull(ctx context.Context, org, repo, branch string) (*models.GitProject, error)
	CreateFile(ctx context.Context, gp models.GitProject, filename string, contents []byte) error
	DeleteFile(ctx context.Context, gp models.GitProject, filename string) error
	CommitAndPush(ctx context.Context, gp models.GitProject, message string) (*models.GitProject, error)
}

type GitCommandDriver struct {
	logger  logr.Logger
	exec    exec.Interface
	baseDir string

	username string
	token    string
}

// TODO: this impl only support https (ssh is not implemented yet)
func NewGitCommandDriver(l logr.Logger, e exec.Interface) (*GitCommandDriver, error) {
	// create basedir
	basedir := models.BaseDir
	if err := os.MkdirAll(basedir, 0755); err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	return &GitCommandDriver{logger: l, exec: e, baseDir: basedir}, nil
}

func (g *GitCommandDriver) WithCredential(username, token string) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	if _, _, err := client.Users.Get(ctx, g.username); err != nil {
		return xerrors.Errorf("%w", err)
	}
	g.username = username
	g.token = token
	return nil
}

func (g *GitCommandDriver) Pull(ctx context.Context, org, repo, branch string) (*models.GitProject, error) {
	downloadDir := filepath.Join(g.baseDir, org, repo)
	// rmdir if already exists
	if _, err := os.Stat(downloadDir); !os.IsNotExist(err) {
		os.RemoveAll(downloadDir)
	}
	// mkdir to $(dirname downloadDir)
	if err := os.MkdirAll(filepath.Dir(downloadDir), 0755); err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	// clone
	url := strings.Join([]string{fmt.Sprintf(baseURL, g.username, g.token), org, repo}, "/") // https://<user>:<token>@github.com/<org>/<repo>
	var stderr bytes.Buffer
	cmd := g.exec.CommandContext(ctx, "git", "clone", "-b", branch, url, downloadDir)
	cmd.SetStderr(&stderr)
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	gp := &models.GitProject{DownlaodDir: downloadDir}
	if err := g.updateHeadCommitSha(ctx, gp); err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	return gp, nil
}

func (g *GitCommandDriver) CreateFile(ctx context.Context, gp models.GitProject, filename string, contents []byte) error {
	fpath := filepath.Join(gp.DownlaodDir, filename)
	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return xerrors.Errorf("%w", err)
	}
	fp, err := os.Create(fpath)
	if err != nil {
		return xerrors.Errorf("%w", err)
	}
	defer fp.Close()
	if _, err := fp.Write(contents); err != nil {
		return xerrors.Errorf("%w", err)
	}
	return nil
}

func (g *GitCommandDriver) DeleteFile(ctx context.Context, gp models.GitProject, filename string) error {
	fpath := filepath.Join(gp.DownlaodDir, filename)
	err := os.RemoveAll(fpath)
	if err != nil {
		return xerrors.Errorf("%w", err)
	}
	return nil
}

func (g *GitCommandDriver) CommitAndPush(ctx context.Context, gp models.GitProject, message string) (*models.GitProject, error) {
	var stdout, stderr bytes.Buffer

	// stage に更新ファイルがない場合早期リターン
	cmd := g.exec.CommandContext(ctx, "git", "status", "-s")
	cmd.SetDir(gp.DownlaodDir)
	stdout = bytes.Buffer{}
	cmd.SetStdout(&stdout)
	stderr = bytes.Buffer{}
	cmd.SetStderr(&stderr)
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	} else if stdout.String() == "" {
		if err := g.updateHeadCommitSha(ctx, &gp); err != nil {
			return nil, xerrors.Errorf("%w", err)
		}
		return &gp, nil
	}

	// add, commit, push
	cmd = g.exec.CommandContext(ctx, "git", "config", "user.name", g.username)
	cmd.SetDir(gp.DownlaodDir)
	stderr = bytes.Buffer{}
	cmd.SetStderr(&stderr)
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	cmd = g.exec.CommandContext(ctx, "git", "config", "user.email", fmt.Sprintf(noreplyEmail, g.username))
	cmd.SetDir(gp.DownlaodDir)
	stderr = bytes.Buffer{}
	cmd.SetStderr(&stderr)
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	cmd = g.exec.CommandContext(ctx, "git", "add", "-A")
	cmd.SetDir(gp.DownlaodDir)
	stderr = bytes.Buffer{}
	cmd.SetStderr(&stderr)
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}

	cmd = g.exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.SetDir(gp.DownlaodDir)
	stderr = bytes.Buffer{}
	cmd.SetStderr(&stderr)
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	cmd = g.exec.CommandContext(ctx, "git", "push", "origin", "HEAD")
	cmd.SetDir(gp.DownlaodDir)
	stderr = bytes.Buffer{}
	cmd.SetStderr(&stderr)
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	if err := g.updateHeadCommitSha(ctx, &gp); err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	return &gp, nil
}

func (g *GitCommandDriver) updateHeadCommitSha(ctx context.Context, gp *models.GitProject) error {
	cmd := g.exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.SetDir(gp.DownlaodDir)
	stdout := bytes.Buffer{}
	cmd.SetStdout(&stdout)
	stderr := bytes.Buffer{}
	cmd.SetStderr(&stderr)
	if err := cmd.Run(); err != nil {
		return xerrors.Errorf(`Error: %v`, stderr.String())
	}
	gp.LatestCommitSha = strings.TrimRight(stdout.String(), "\n")
	return nil
}

func (g *GitCommandDriver) HashLogs(ctx context.Context, gp models.GitProject, hash1, hash2 string) ([]string, error) {
	cmd := g.exec.CommandContext(ctx, "git", "log", "--format=%H",
		fmt.Sprintf("%s...%s", hash1, hash2),
	)
	cmd.SetDir(gp.DownlaodDir)
	cmd.SetDir(gp.DownlaodDir)
	stdout := bytes.Buffer{}
	cmd.SetStdout(&stdout)
	stderr := bytes.Buffer{}
	cmd.SetStderr(&stderr)
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	return strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n"), nil
}
