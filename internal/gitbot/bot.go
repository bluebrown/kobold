package gitbot

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/bluebrown/kobold/kobold/config"
)

type PullRequester interface {
	Open(ctx context.Context, src, target, title, description string) error
}

type GitTransporter interface {
	Path() string
	URL() string
	Auth() *url.Userinfo
	Refresh(ctx context.Context, branch string) error
	CheckoutBranch(ctx context.Context, branch string) error
	AddCommitPush(ctx context.Context, branch, title, description string) error
}

func NewRepo(tranport GitTransporter, provider config.GitProvider) *repo {
	if provider == "" {
		provider = config.InferGitProvider(tranport.URL())
	}
	return &repo{
		lock:      sync.Mutex{},
		transport: tranport,
		provider:  provider,
	}
}

type repo struct {
	transport GitTransporter
	provider  config.GitProvider
	lock      sync.Mutex
}

type Repos map[string]*repo

func (r *repo) Transaction(fn func(path string, transport GitTransporter) error) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := fn(r.transport.Path(), r.transport)
	return err
}

func (r *repo) Provider() config.GitProvider {
	return r.provider
}

func (r *repo) URL() string {
	return r.transport.URL()
}

func (r *repo) Auth() *url.Userinfo {
	return r.transport.Auth()
}

type GitBot struct {
	repo   *repo
	branch string
	pr     PullRequester
	logger zerolog.Logger
}

func NewGitbot(name string, repo *repo, branch string, pr PullRequester) *GitBot {
	return &GitBot{
		repo:   repo,
		branch: branch,
		pr:     pr,
		logger: log.With().Str("repo", name).Logger(),
	}
}

type DoCallback func(ctx context.Context, dir string) (title string, body string, changed bool, err error)

func (bot *GitBot) Do(ctx context.Context, callback DoCallback) error {
	return bot.repo.Transaction(func(path string, transport GitTransporter) error {
		// discard all changes, checkout the branch and fetch new changes
		if err := transport.Refresh(ctx, bot.branch); err != nil {
			return err
		}

		// execute the callback for this actions
		title, description, changed, err := callback(ctx, path)
		if err != nil {
			return err
		}

		if !changed {
			return nil
		}

		// wehen there is no pull requester, make a commit to the given branch and push
		if bot.pr == nil {
			err := transport.AddCommitPush(ctx, bot.branch, title, description)
			if err != nil {
				return err
			}
			if changed {
				bot.logger.Info().Str("base", bot.branch).Str("action", "commit/push").Msg("change detected")
			}
			return nil
		}

		// if there is a pull requester, create a new branch with timestamp
		// and create a pull request to the given branch
		branch := fmt.Sprintf("kobold/%d", time.Now().Unix())

		if err := transport.CheckoutBranch(ctx, branch); err != nil {
			return err
		}

		err = transport.AddCommitPush(ctx, fmt.Sprintf("kobold/%d", time.Now().Unix()), title, description)
		if err != nil {
			return err
		}

		bot.logger.Info().Str("head", branch).Str("base", bot.branch).Str("action", "commit/push/pr").Msg("changed detected")
		return bot.pr.Open(ctx, branch, bot.branch, title, description)

	})
}
