//go:build gitexec

package transport

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"

	"github.com/bluebrown/kobold/internal/gitbot"
	"github.com/rs/zerolog/log"
)

// tag is used in the version flag output
var Tag = "gitexec"

// TODO: GIT_SSH_COMMAND='ssh -i private_key_file -o IdentitiesOnly=yes'
// ref: https://stackoverflow.com/a/29754018/9208887

// ensure interface compliance
var _ gitbot.GitTransporter = (*Transport)(nil)

type Transport struct {
	url  string
	path string
	auth *url.Userinfo
}

func New(repoUri string, path string, auth *url.Userinfo) (*Transport, error) {
	parsedUrl, err := url.Parse(repoUri)
	if err != nil {
		return nil, err
	}
	if auth != nil {
		parsedUrl.User = auth
	}
	repoUri = parsedUrl.String()

	if err := clone(context.Background(), repoUri, path); err != nil && err != ErrAlreadyExists {
		return nil, err
	}
	log.Trace().Str("uri", repoUri).Msg("setting identify")
	if err := setIdentity(context.Background(), path, "Kobold", "kobold@noreply.sh"); err != nil {
		return nil, err
	}
	return &Transport{
		url:  repoUri,
		path: path,
		auth: auth,
	}, nil
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
	if err := checkout(ctx, t.path, branch); err != nil {
		return fmt.Errorf("could not checkout branch %s: %w", branch, err)
	}
	if err := fetch(ctx, t.path, branch); err != nil {
		return fmt.Errorf("could not fetch branch %s: %w", branch, err)
	}
	if err := resetHard(ctx, t.path, branch); err != nil {
		return fmt.Errorf("could not reset branch %s: %w", branch, err)
	}
	return nil
}

func (t *Transport) CheckoutBranch(ctx context.Context, branch string) error {
	if err := checkoutBranch(ctx, t.path, branch); err != nil {
		return fmt.Errorf("could not checkout branch %s: %w", branch, err)
	}
	return nil
}

func (t *Transport) AddCommitPush(ctx context.Context, branch, title, description string) (bool, error) {
	ok, err := changed(ctx, t.path)
	if err != nil {
		return false, fmt.Errorf("could not check for changes: %w", err)
	}
	if !ok {
		return false, nil
	}
	if err := add(ctx, t.path, "."); err != nil {
		return true, fmt.Errorf("could not stage changes: %w", err)
	}
	if err := commit(ctx, t.path, title, description); err != nil {
		return true, fmt.Errorf("could not commit staged changes: %w", err)
	}

	if err := push(ctx, t.path, branch); err != nil {
		return true, fmt.Errorf("could not push: %w", err)
	}
	return true, nil
}

var (
	ErrAlreadyExists = errors.New("repo already exists")
)

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func do(ctx context.Context, args ...string) error {
	b, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(b))
	}
	return nil
}

func clone(ctx context.Context, upstreamRepo, path string) error {
	if exists(path) {
		return ErrAlreadyExists
	}
	return do(ctx, "clone", upstreamRepo, path)
}

func pull(ctx context.Context, path, branch string) error {
	return do(ctx, "-C", path, "pull", "origin", branch)
}

func add(ctx context.Context, path, relSrc string) error {
	return do(ctx, "-C", path, "add", relSrc)
}

func commit(ctx context.Context, path, title, description string) error {
	return do(ctx, "-C", path, "commit", "-m", title, "-m", description)
}

func push(ctx context.Context, path, branch string) error {
	return do(ctx, "-C", path, "push", "origin", branch)
}

func changed(ctx context.Context, path string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "status", "--porcelain")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%w: %s", err, string(b))
	}
	return len(b) > 0, nil
}

func setIdentity(ctx context.Context, path, name, email string) error {
	if err := do(ctx, "-C", path, "config", "user.name", name); err != nil {
		return err
	}
	return do(ctx, "-C", path, "config", "user.email", email)
}

func branch(ctx context.Context, path, branch string) error {
	return do(ctx, "-C", path, "branch", branch)
}

func checkout(ctx context.Context, path, branch string) error {
	return do(ctx, "-C", path, "checkout", branch)
}

func checkoutBranch(ctx context.Context, path, branch string) error {
	return do(ctx, "-C", path, "checkout", "-b", branch)
}

func fetch(ctx context.Context, path, branch string) error {
	return do(ctx, "-C", path, "fetch", "origin", branch)
}

func resetHard(ctx context.Context, path, branch string) error {
	return do(ctx, "-C", path, "reset", "--hard", fmt.Sprintf("origin/%s", branch))
}
