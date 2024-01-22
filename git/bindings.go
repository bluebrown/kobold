package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// initialite the git repo at dir and set the remote origin to uri
func Init(ctx context.Context, dir, uri string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %q: %w", dir, err)
	}

	if _, err := run(ctx, dir, "init"); err != nil {
		return fmt.Errorf("git init: %w", err)
	}

	if _, err := run(ctx, dir, "remote", "add", "origin", uri); err != nil {
		return fmt.Errorf("git remote add origin %q: %w", uri, err)
	}

	return nil
}

// fetch the given refs from origin with depth 1
func FetchShallow(ctx context.Context, dir string, refs ...string) error {
	refs = unique(refs)
	args := []string{"fetch", "--depth", "1", "origin"}
	clean := make([]string, 0, len(refs))
	for _, ref := range refs {
		clean = append(clean, fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", ref, ref))
	}
	args = append(args, clean...)
	_, err := run(ctx, dir, args...)
	return err
}

// perform init and fetch in one step but only init if the repo doesn't exist.
// otherwise, update its fetch refs and re-fetch with depth 1
func Ensure(ctx context.Context, dir, uri string, refs ...string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := Init(ctx, dir, uri); err != nil {
				return fmt.Errorf("git init: %w", err)
			}
		} else {
			return fmt.Errorf("stat %q: %w", dir, err)
		}
	}

	if err := FetchShallow(ctx, dir, refs...); err != nil {
		return fmt.Errorf("git fetch: %w", err)
	}

	return nil
}

// switch to ref. This should be called after Ensure
// to get a writeable branch
func Switch(ctx context.Context, dir, ref string) error {
	_, err := run(ctx, dir, "checkout", ref)
	return err
}

// create a new branch and check it out
func CheckoutB(ctx context.Context, dir, ref string) error {
	_, err := run(ctx, dir, "checkout", "-b", ref)
	return err
}

// add all files in given dir to the index
func AddRoot(ctx context.Context, dir string) error {
	_, err := run(ctx, dir, "add", ".")
	return err
}

// commit the index with the given message
func Commit(ctx context.Context, dir, msg string) error {
	_, err := run(ctx, dir, "commit", "-m", msg)
	return err
}

// push the given refs to origin
func Push(ctx context.Context, dir string, refs ...string) error {
	args := []string{"push", "origin"}
	args = append(args, refs...)
	_, err := run(ctx, dir, args...)
	return err
}

// perform add, commit, and push in one step, on the current branch
func Publish(ctx context.Context, dir, ref, msg string) error {
	if err := AddRoot(ctx, dir); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	if err := Commit(ctx, dir, msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	if err := Push(ctx, dir, ref); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	return nil
}

func run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", err, string(b), strings.Join(args, " "))
	}
	return b, nil
}
