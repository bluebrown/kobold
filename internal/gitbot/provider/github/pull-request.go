package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// PR
// ref: https://docs.github.com/en/rest/pulls/pulls#create-a-pull-request

// TODO: implement
// REVIEW
// ref: https://docs.github.com/en/rest/pulls/review-requests#request-reviewers-for-a-pull-request

/*
Example Payload:

	{
		"title": "Amazing new feature",
		"body": "Please pull these awesome changes in!",
		"head":"octocat:new-feature",
		"base":"master"
	}
*/
type PullRequestPayload struct {
	Head                string `json:"head,omitempty"`
	Base                string `json:"base,omitempty"`
	Title               string `json:"title,omitempty"`
	Body                string `json:"body,omitempty"`
	MaintainerCanModify bool   `json:"maintainer_can_modify,omitempty"`
	Draft               bool   `json:"draft,omitempty"`
	Issue               *int   `json:"issue,omitempty"`
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type pullRequestClient struct {
	client httpClient
	owner  string
	repo   string
	auth   *url.Userinfo
}

// POST https://api.github.com/repos/OWNER/REPO/pulls
func (pr pullRequestClient) Open(ctx context.Context, src, target, title, description string) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&PullRequestPayload{
		Head:                src,
		Base:                target,
		Title:               title,
		Body:                description,
		MaintainerCanModify: true,
		Draft:               false,
	}); err != nil {
		return err
	}
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", pr.owner, pr.repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return err
	}
	req.URL.User = pr.auth
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := pr.client.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("expected status %d but got %d", http.StatusCreated, res.StatusCode)
	}
	return nil
}

func NewPrClient(upstreamRepo string, auth *url.Userinfo, client httpClient) (*pullRequestClient, error) {
	owner, repo, err := getOwnerRepo(upstreamRepo)
	if err != nil {
		return nil, err
	}

	if client == nil {
		client = http.DefaultClient
	}

	return &pullRequestClient{
		client: client,
		owner:  owner,
		repo:   repo,
		auth:   auth,
	}, nil
}

// https://github.com/bluebrown/kobold-test
// git@github.com:bluebrown/kobold-test.git
func getOwnerRepo(u string) (owner, repo string, err error) {
	u = strings.TrimPrefix(u, "git@github.com:")
	u = strings.TrimSuffix(u, "/")
	u = strings.TrimSuffix(u, ".git")
	parts := strings.Split(u, "/")
	if len(parts) < 2 {
		return "", "", errors.New("invalid url")
	}
	owner = parts[len(parts)-2]
	repo = parts[len(parts)-1]
	return owner, repo, err
}
