package azure

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

var (
	ErrPullRequestAlreadyExists = errors.New("an active pull request for the source and target branch already exists")
)

/*
Example Payload:

	{
		"sourceRefName": "refs/heads/npaulk/my_work",
		"targetRefName": "refs/heads/new_feature",
		"title": "A new feature",
		"description": "Adding a new feature",
		"reviewers": [
		  {
			"id": "d6245f20-2af8-44f4-9451-8107cb2767db"
		  }
		]
	  }
*/
type PullRequestPayload struct {
	SourceRefName string                `json:"sourceRefName,omitempty"`
	TargetRefName string                `json:"targetRefName,omitempty"`
	Title         string                `json:"title,omitempty"`
	Description   string                `json:"description,omitempty"`
	Reviewers     []PullRequestReviewer `json:"reviewers,omitempty"`
}

type PullRequestReviewer struct {
	Id string `json:"id,omitempty"`
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type pullRequestClient struct {
	client  httpClient
	org     string
	project string
	repoId  string
	auth    *url.Userinfo
}

// POST https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}/pullrequests?api-version=7.0
func (pr pullRequestClient) Open(ctx context.Context, src, target, title, description string) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&PullRequestPayload{
		SourceRefName: "refs/heads/" + src,
		TargetRefName: "refs/heads/" + target,
		Title:         title,
		Description:   description,
	}); err != nil {
		return err
	}
	u := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/git/repositories/%s/pullrequests?api-version=7.0", pr.org, pr.project, pr.repoId)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return err
	}
	// TODO:check is header is set, otherwise set header ourself
	req.URL.User = pr.auth
	req.Header.Set("Content-Type", "application/json")

	res, err := pr.client.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()

	if res.StatusCode == http.StatusConflict {
		return ErrPullRequestAlreadyExists
	}

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("expected status %d but got %d", http.StatusCreated, res.StatusCode)
	}
	return nil
}

func NewPrClient(upstreamRepo string, auth *url.Userinfo, client httpClient) (*pullRequestClient, error) {
	org, project, repoId, err := getOrgProjRepo(upstreamRepo)
	if err != nil {
		return nil, err
	}

	if client == nil {
		client = http.DefaultClient
	}

	return &pullRequestClient{
		client:  client,
		org:     org,
		project: project,
		repoId:  repoId,
		auth:    auth,
	}, nil
}

// https://<org>@dev.azure.com/<org>/<project>/_git/<repoId>
// git@ssh.dev.azure.com:v3/<org>/<project>/<repoId>
func getOrgProjRepo(u string) (org, proj, repo string, err error) {
	u = strings.TrimSuffix(u, ".git")
	parts := strings.Split(u, "/")
	if strings.HasPrefix(u, "git@ssh") {
		if len(parts) != 4 {
			return "", "", "", fmt.Errorf("invalid url")
		}
		return parts[1], parts[2], parts[3], nil
	}
	if len(parts) != 7 {
		return "", "", "", fmt.Errorf("invalid url")
	}
	return parts[3], parts[4], parts[6], nil
}
