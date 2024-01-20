package confix

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/bluebrown/kobold/config"
	"github.com/bluebrown/kobold/config/confix/old"
	"github.com/bluebrown/kobold/git"
)

var prm = map[string]string{
	"github": "builtin.github-pr@v1",
	"azure":  "builtin.ado-pr@v1",
	"gitea":  "builtin.gitea-pr@v1",
}

var dm = map[string]string{
	"generic":      "builtin.lines@v1",
	"acr":          "builtin.distribution@v1",
	"distribution": "builtin.distribution@v1",
	"dockerhub":    "builtin.dockerhub@v1",
}

var gitconfig = []byte(`
[credential]
	helper = store --file ~/.git-credentials
`)

func MakeConfig(v1 *old.NormalizedConfig) (*config.Config, error) {
	v2 := config.Config{
		Version: "2",
	}

	for _, ep := range v1.Endpoints {
		d, ok := dm[string(ep.Type)]
		if !ok {
			fmt.Printf("[WARN] endpoint=%q: unsupported decoder %q, using default!\n", string(ep.Type), ep.Name)
		}

		v2.Channels = append(v2.Channels, config.Channel{
			Name:    ep.Name,
			Decoder: d,
		})

	}

	for _, sub := range v1.Subscriptions {
		repo := find(v1.Repositories, func(r old.RepositorySpec) bool { return r.Name == sub.RepositoryRef.Name })

		isPr := sub.Strategy == "pull-request"

		cc := make([]string, len(sub.EndpointRefs))
		for i, e := range sub.EndpointRefs {
			cc[i] = e.Name
		}

		if len(sub.Scopes) == 0 {
			sub.Scopes = []string{""}
		}

		for _, scope := range sub.Scopes {
			if scope != "" && !strings.HasPrefix(scope, "/") {
				fmt.Printf("[WARN] sub=%q repo=%q scope=%q: unsupported scope, skipping!\n",
					sub.Name, repo.Name, scope)
				continue
			}

			uri := git.PackageURI{}
			uri.UnmarshalText([]byte(fmt.Sprintf("%s@%s%s", repo.URL, sub.Branch, scope)))

			scope = strings.ReplaceAll(scope, "/", "-")
			scope = strings.TrimSuffix(scope, "-")

			if repo.Provider == "" {
				repo.Provider = old.InferGitProvider(repo.URL)
			}

			v2.Pipelines = append(v2.Pipelines, config.Pipeline{
				Name:       sub.Name + scope,
				RepoURI:    uri,
				DestBranch: tern(isPr, "kobold", ""),
				PostHook:   tern(isPr, prm[string(repo.Provider)], ""),
				Channels:   cc,
			})
		}
	}

	return &v2, nil
}

func MakeGitCredentials(v1 *old.NormalizedConfig) (string, error) {
	var buf bytes.Buffer

	seen := map[string]struct{}{}

	for _, repo := range v1.Repositories {
		if repo.Username == "" || repo.Password == "" {
			continue
		}

		u, err := url.Parse(repo.URL)
		if err != nil {
			warn("nvalid url %q: %v\n", repo.URL, err)
			continue
		}

		pw := repo.Password
		us := repo.Username

		if v := u.User.Username(); v != "" {
			warn("repo=%q: username already set to %s\n", repo.Name, v)
			us = v
		}

		if v, ok := u.User.Password(); ok {
			warn("repo=%q: password already set\n", repo.Name)
			pw = v
		}

		u.User = url.UserPassword(us, pw)
		u.Path = ""
		key := u.String()

		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			fmt.Fprintf(&buf, "%s\n", key)
		}

	}

	return buf.String(), nil
}

func MakeGitConfig() []byte {
	return gitconfig
}

func find[T any](s []T, f func(T) bool) T {
	for _, v := range s {
		if f(v) {
			return v
		}
	}
	return *new(T)
}

func tern[T any](cond bool, yes, no T) T {
	if cond {
		return yes
	}
	return no
}

var seenw = map[string]struct{}{}

func warn(msg string, args ...interface{}) {
	if _, ok := seenw[msg]; !ok {
		fmt.Printf("[WARNING] "+msg+"\n", args...)
		seenw[msg] = struct{}{}
	}
}
