package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudnativedaysjp/reviewapp-operator/models"
	"github.com/go-logr/logr"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

const (
	baseURL      = `https://%s:%s@github.com`
	noreplyEmail = `%s@users.noreply.github.com`
)

type GitCommandDriver struct {
	logger logr.Logger

	baseDir  string
	username string
	token    string
}

// TODO: this impl only support https (ssh is not implemented yet)
func NewGitCommandDriver(l logr.Logger) (*GitCommandDriver, error) {
	// create basedir
	basedir := models.BaseDir
	if err := os.MkdirAll(basedir, 0755); err != nil {
		return nil, err
	}

	return &GitCommandDriver{logger: l, baseDir: basedir}, nil
}

func (g *GitCommandDriver) WithCredential(username, token string) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	if _, _, err := client.Users.Get(ctx, g.username); err != nil {
		return err
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
		return nil, err
	}
	// clone
	url := strings.Join([]string{fmt.Sprintf(baseURL, g.username, g.token), org, repo}, "/") // https://<user>:<token>@github.com/<org>/<repo>
	cmd := exec.CommandContext(ctx, "git", "clone", "-b", branch, url, downloadDir)
	if out, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf(`Error: %v: %v`, err, out)
	}
	gp := &models.GitProject{DownlaodDir: downloadDir}
	if err := g.updateHeadCommitSha(ctx, gp); err != nil {
		return nil, err
	}
	return gp, nil
}

func (g *GitCommandDriver) CreateFile(ctx context.Context, gp models.GitProject, filename string, contents []byte) error {
	fpath := filepath.Join(gp.DownlaodDir, filename)
	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return err
	}
	fp, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer fp.Close()
	if _, err := fp.Write(contents); err != nil {
		return err
	}
	return nil
}

func (g *GitCommandDriver) DeleteFile(ctx context.Context, gp models.GitProject, filename string) error {
	fpath := filepath.Join(gp.DownlaodDir, filename)
	return os.RemoveAll(fpath)
}

func (g *GitCommandDriver) CommitAndPush(ctx context.Context, gp models.GitProject, message string) (*models.GitProject, error) {
	// stage に更新ファイルがない場合早期リターン
	cmd := exec.CommandContext(ctx, "git", "status", "-s")
	cmd.Dir = gp.DownlaodDir
	if out, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf(`Error: %v: %v`, err, out)
	} else if string(out) == "" {
		if err := g.updateHeadCommitSha(ctx, &gp); err != nil {
			return nil, err
		}
		return &gp, nil
	}

	// add, commit, push
	cmd = exec.CommandContext(ctx, "git", "config", "user.name", g.username)
	cmd.Dir = gp.DownlaodDir
	if out, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf(`Error: %v: %v`, err, out)
	}
	cmd = exec.CommandContext(ctx, "git", "config", "user.email", fmt.Sprintf(noreplyEmail, g.username))
	cmd.Dir = gp.DownlaodDir
	if out, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf(`Error: %v: %v`, err, out)
	}
	cmd = exec.CommandContext(ctx, "git", "add", "-A")
	cmd.Dir = gp.DownlaodDir
	if out, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf(`Error: %v: %v`, err, out)
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = gp.DownlaodDir
	if out, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf(`Error: %v: %v`, err, out)
	}
	cmd = exec.CommandContext(ctx, "git", "push", "origin", "HEAD")
	cmd.Dir = gp.DownlaodDir
	if out, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf(`Error: %v: %v`, err, out)
	}
	if err := g.updateHeadCommitSha(ctx, &gp); err != nil {
		return nil, err
	}
	return &gp, nil
}

func (g *GitCommandDriver) updateHeadCommitSha(ctx context.Context, gp *models.GitProject) error {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = gp.DownlaodDir
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(`Error: %v: %v`, err, out)
	}
	gp.LatestCommitSha = strings.TrimRight(string(out), "\n")
	return nil
}

func (g *GitCommandDriver) HashLogs(ctx context.Context, gp models.GitProject, hash1, hash2 string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "--format=%H",
		fmt.Sprintf("%s...%s", hash1, hash2),
	)
	cmd.Dir = gp.DownlaodDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf(`Error: %v: %v`, err, out)
	}
	return strings.Split(strings.TrimRight(string(out), "\n"), "\n"), nil
}
