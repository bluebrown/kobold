package server

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"

	"github.com/bluebrown/kobold/internal/events"
	"github.com/bluebrown/kobold/internal/events/acr"
	"github.com/bluebrown/kobold/internal/events/dockerhub"
	"github.com/bluebrown/kobold/internal/events/generic"
	"github.com/bluebrown/kobold/internal/gitbot"
	"github.com/bluebrown/kobold/internal/gitbot/provider/azure"
	"github.com/bluebrown/kobold/internal/gitbot/provider/github"
	"github.com/bluebrown/kobold/internal/gitbot/transport"
	"github.com/bluebrown/kobold/internal/krm"
	"github.com/bluebrown/kobold/internal/registry"
	"github.com/bluebrown/kobold/kobold/config"
)

type generator struct {
	dataDir          string
	useK8sChain      bool
	defaultRegistry  string
	imagerefTemplate string
}

func (g generator) Generate(conf *config.NormalizedConfig) (http.Handler, error) {
	// initialize all repositories
	repos := make(gitbot.Repos, 0)
	for _, r := range conf.Repositories {
		log.Info().Str("repo", r.Name).Str("url", r.URL).Msg("setup repo")
		transport, err := transport.New(
			r.URL,
			filepath.Join(g.dataDir, r.Name),
			url.UserPassword(r.Username, r.Password),
		)
		if err != nil {
			return nil, fmt.Errorf("git transport: %s: %w", r.URL, err)
		}
		if _, ok := repos[r.Name]; ok {
			log.Warn().Str("repo", r.Name).Msg("repo already registered, skipping")
			continue
		}
		repos[r.Name] = gitbot.NewRepo(transport, r.Provider)
	}

	// set up the channels for the subscriptions
	subsMap := make(map[string][]chan events.PushData)
	for _, sub := range conf.Subscriptions {
		repo, ok := repos[sub.RepositoryRef.Name]
		if !ok {
			log.Warn().Str("sub", sub.Name).Str("repo", sub.RepositoryRef.Name).Msg("repo does not exist, skipping")
			continue
		}

		log.Info().Str("sub", sub.Name).Str("branch", sub.Branch).Str("strategy", string(sub.Strategy)).Strs("scopes", sub.Scopes).Msg("setup sub")

		var (
			prClient gitbot.PullRequester
			err      error
		)

		// use this retryable client for pull requests
		retryClient := retryablehttp.NewClient()
		retryClient.Logger = nil
		retryClient.RetryMax = 3
		standardClient := retryClient.StandardClient()

		switch sub.Strategy {
		case config.StrategyCommit:
			prClient = nil
		case config.StrategyPullRequest:
			log.Debug().Str("sub", sub.Name).Str("provider", string(repo.Provider())).Msg("setup pull requests")
			switch repo.Provider() {
			case "":
				return nil, fmt.Errorf("using pull-requests requires a known provider")
			case config.ProviderGithub:
				prClient, err = github.NewPrClient(repo.URL(), repo.Auth(), standardClient)
			case config.ProviderAzure:
				prClient, err = azure.NewPrClient(repo.URL(), repo.Auth(), standardClient)
			default:
				return nil, fmt.Errorf("provider %s not supported for pull-requests", repo.Provider())
			}
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported strategy: %s", sub.Strategy)
		}

		bot := gitbot.NewGitbot(sub.Name, repo, sub.Branch, prClient)

		renderer := krm.NewRenderer(
			krm.WithScopes(sub.Scopes),
			krm.WithSelector(krm.NewSelector(conf.Resolvers, sub.FileAssociations)),
			// these 2 could be part of the subscription config
			// for now, they will be global for all configs
			krm.WithDefaultRegistry(g.defaultRegistry),
			krm.WithImagerefTemplate(g.imagerefTemplate),
		)

		// TODO: check if sub with given name already exists and warn user
		subChan := NewSubscriber(
			sub.Name,
			bot,
			renderer,
			gitbot.NewTemplateCommitMessenger(conf.CommitMessage.Title, conf.CommitMessage.Description),
		)

		for _, ef := range sub.EndpointRefs {
			subsMap[ef.Name] = append(subsMap[ef.Name], subChan)
		}
	}

	mux := http.NewServeMux()

	log.Trace().
		Str("namespace", conf.RegistryAuth.Namespace).
		Str("serviceAccount", conf.RegistryAuth.ServiceAccount).
		Interface("imagePullSecrets", conf.RegistryAuth.ImagePullSecrets).
		Msg("using registryAuth")

	keys, err := registry.NewKeys(g.useK8sChain, conf.RegistryAuth)
	if err != nil {
		return nil, err
	}

	for _, endpoint := range conf.Endpoints {
		log.Info().Str("endpoint", endpoint.Name).Str("path", endpoint.Path).Msg("setup endpoint")
		var ph events.PayloadHandler
		switch endpoint.Type {
		case config.EndpointTypeGeneric:
			ph = generic.NewPayloadHandler()
		case config.EndpointTypeACR:
			ph = acr.NewPayloadHandler()
		case config.EndpointTypeDockerhub:
			ph = dockerhub.NewPayloadHandler(registry.NewDigestFetcher(g.defaultRegistry, keys))
		default:
			return nil, fmt.Errorf("unsupported endpoint type: %s", endpoint.Type)
		}
		mux.Handle(endpoint.Path, RequireHeaders(
			endpoint.RequiredHeaders,
			NewPushWebhook(endpoint.Name, subsMap[endpoint.Name], ph),
		))
	}

	return mux, nil
}
