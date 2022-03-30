package gitcommand

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
	"k8s.io/utils/exec"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

const (
	baseURL      = `https://%s:%s@github.com`
	noreplyEmail = `%s@users.noreply.github.com`
	BaseDir      = "/tmp"
)

type Git struct {
	logger  logr.Logger
	exec    exec.Interface
	baseDir string

	username string
	token    string
}

// TODO: this impl only support https (ssh is not implemented yet)
func NewGit(l logr.Logger, e exec.Interface) (*Git, error) {
	// create basedir
	basedir := BaseDir
	if err := os.MkdirAll(basedir, 0755); err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	return &Git{logger: l, exec: e, baseDir: basedir}, nil
}

func (g *Git) WithCredential(credential models.GitCredential) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: credential.Token()},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	if _, _, err := client.Users.Get(ctx, g.username); err != nil {
		return xerrors.Errorf("%w", err)
	}
	g.username = credential.Username()
	g.token = credential.Token()
	return nil
}

func (g *Git) ForceClone(ctx context.Context, infraTarget models.InfraRepoTarget) (models.InfraRepoLocalDir, error) {
	downloadDir := filepath.Join(g.baseDir, infraTarget.Organization, infraTarget.Repository)
	// rmdir if already exists
	if _, err := os.Stat(downloadDir); !os.IsNotExist(err) {
		os.RemoveAll(downloadDir)
	}
	// mkdir to $(dirname downloadDir)
	if err := os.MkdirAll(filepath.Dir(downloadDir), 0755); err != nil {
		return models.InfraRepoLocalDir{}, xerrors.Errorf("%w", err)
	}
	// clone
	url := strings.Join([]string{fmt.Sprintf(baseURL, g.username, g.token), infraTarget.Organization, infraTarget.Repository}, "/") // https://<user>:<token>@github.com/<org>/<repo>
	_, stderr, err := g.runCommand(ctx, "", "git", "clone", "-b", infraTarget.Branch, url, downloadDir)
	if err != nil {
		return models.InfraRepoLocalDir{}, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	gp := models.NewInfraRepoLocal(downloadDir)
	gp, err = g.updateHeadCommitSha(ctx, gp)
	if err != nil {
		return models.InfraRepoLocalDir{}, xerrors.Errorf("%w", err)
	}
	return gp, nil
}

func (g *Git) CreateFiles(ctx context.Context, gp models.InfraRepoLocalDir, files ...models.File) error {
	for _, f := range files {
		fpath := filepath.Join(gp.BaseDir(), f.Filepath)
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return xerrors.Errorf("%w", err)
		}
		fp, err := os.Create(fpath)
		if err != nil {
			return xerrors.Errorf("%w", err)
		}
		defer fp.Close()
		if _, err := fp.Write(f.Content); err != nil {
			return xerrors.Errorf("%w", err)
		}
	}
	return nil
}

func (g *Git) DeleteFiles(ctx context.Context, gp models.InfraRepoLocalDir, files ...models.File) error {
	for _, f := range files {
		fpath := filepath.Join(gp.BaseDir(), f.Filepath)
		err := os.RemoveAll(fpath)
		if err != nil {
			return xerrors.Errorf("%w", err)
		}
	}
	return nil
}

func (g *Git) CommitAndPush(ctx context.Context, gp models.InfraRepoLocalDir, message string) (*models.InfraRepoLocalDir, error) {
	// stage に更新ファイルがない場合早期リターン
	stdout, stderr, err := g.runCommand(ctx, gp.BaseDir(), "git", "status", "-s")
	if err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	} else if stdout.String() == "" {
		gp, err := g.updateHeadCommitSha(ctx, gp)
		if err != nil {
			return nil, xerrors.Errorf("%w", err)
		}
		return &gp, nil
	}

	// add, commit, push
	stdout, stderr, err = g.runCommand(ctx, gp.BaseDir(), "git", "config", "user.name", g.username)
	if err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	stdout, stderr, err = g.runCommand(ctx, gp.BaseDir(), "git", "config", "user.email", fmt.Sprintf(noreplyEmail, g.username))
	if err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	stdout, stderr, err = g.runCommand(ctx, gp.BaseDir(), "git", "add", "-A")
	if err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	stdout, stderr, err = g.runCommand(ctx, gp.BaseDir(), "git", "commit", "-m", message)
	if err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	stdout, stderr, err = g.runCommand(ctx, gp.BaseDir(), "git", "push", "origin", "HEAD")
	if err != nil {
		return nil, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	gp, err = g.updateHeadCommitSha(ctx, gp)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	return &gp, nil
}

func (g *Git) updateHeadCommitSha(ctx context.Context, gp models.InfraRepoLocalDir) (models.InfraRepoLocalDir, error) {
	stdout, stderr, err := g.runCommand(ctx, gp.BaseDir(), "git", "rev-parse", "HEAD")
	if err != nil {
		return models.InfraRepoLocalDir{}, xerrors.Errorf(`Error: %v`, stderr.String())
	}
	gp = gp.SetLatestCommitSha(strings.TrimRight(stdout.String(), "\n"))
	return gp, nil
}

func (g *Git) runCommand(ctx context.Context, basedir string, cmd string, args ...string) (bytes.Buffer, bytes.Buffer, error) {
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cc := g.exec.CommandContext(ctx, cmd, args...)
	if basedir != "" {
		cc.SetDir(basedir)
	}
	cc.SetStdout(&stdout)
	cc.SetStderr(&stderr)
	if err := cc.Run(); err != nil {
		return stdout, stderr, err
	}
	return stdout, stderr, nil
}
