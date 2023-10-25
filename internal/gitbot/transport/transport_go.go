//go:build !gitexec

package transport

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/bluebrown/kobold/internal/gitbot"
)

func init() {
	// this is used to remove multiack from the defaults
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}
}

// tag is used in the version flag output
var Tag = "gitgo"

// ensure interface compliance
var _ gitbot.GitTransporter = (*Transport)(nil)

type Transport struct {
	url   string
	path  string
	auth  *url.Userinfo
	repo  *git.Repository
	gauth *http.BasicAuth
}

func New(repoUri string, path string, auth *url.Userinfo) (*Transport, error) {
	pw, _ := auth.Password()
	t := &Transport{
		url:  repoUri,
		path: path,
		auth: auth,
		gauth: &http.BasicAuth{
			Username: auth.Username(),
			Password: pw,
		},
	}

	if err := os.RemoveAll(t.path); err != nil {
		return t, err
	}

	repo, err := git.PlainCloneContext(context.Background(), t.path, false, &git.CloneOptions{
		Auth:       t.gauth,
		URL:        t.url,
		RemoteName: "origin",
		// SingleBranch: true,
	})
	if err != nil {
		return t, err
	}

	t.repo = repo

	return t, nil
}

func (t *Transport) URL() string {
	return t.url
}

func (t *Transport) Path() string {
	return t.path
}

func (t *Transport) Auth() *url.Userinfo {
	return t.auth
}

func (t *Transport) Refresh(ctx context.Context, branch string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := os.RemoveAll(t.path); err != nil {
		return err
	}

	repo, err := git.PlainCloneContext(ctx, t.path, false, &git.CloneOptions{
		Auth:          t.gauth,
		URL:           t.url,
		RemoteName:    "origin",
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return fmt.Errorf("could not refesh: failed to clone: %w", err)
	}

	t.repo = repo

	return nil
}

func (t *Transport) CheckoutBranch(ctx context.Context, branch string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	w, err := t.repo.Worktree()
	if err != nil {
		return fmt.Errorf("could not checkout branch: %w", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Create: true,
		Keep:   true,
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
	})
	if err != nil {
		return fmt.Errorf("could not checkout branch: %w", err)
	}

	return nil
}

func (t *Transport) AddCommitPush(ctx context.Context, branch, title, description string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	w, err := t.repo.Worktree()
	if err != nil {
		return fmt.Errorf("get worktree: %w", err)
	}

	s, err := w.Status()
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	if s.IsClean() {
		return ErrNoChange
	}

	_, err = w.Add(".")
	if err != nil {
		return fmt.Errorf("add changes: %w", err)
	}

	_, err = w.Commit(fmt.Sprintf("%s\n\n%s", title, description), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Kobold",
			Email: "kobold@noreply.sh",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	err = t.repo.PushContext(ctx, &git.PushOptions{
		RemoteURL:  t.url,
		RemoteName: "origin",
		Auth:       t.gauth,
	})

	if err != nil {
		return fmt.Errorf("push: %w", err)
	}

	return nil
}
