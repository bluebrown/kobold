package confix

import (
	"github.com/bluebrown/kobold/config"
	"github.com/bluebrown/kobold/config/confix/old"
	"github.com/bluebrown/kobold/git"
)

func MakeConfig(oldConf *old.Config) (*config.Config, error) {
	newConf := config.Config{}
	newConf.Version = oldConf.Version

	for _, oldChannel := range oldConf.Channels {
		newConf.Channels = append(newConf.Channels, config.Channel(oldChannel))
	}

	for _, oldPipeline := range oldConf.Pipelines {
		newConf.Pipelines = append(newConf.Pipelines, config.Pipeline{
			Name:       oldPipeline.Name,
			RepoURI:    git.PackageURI(oldPipeline.RepoURI),
			DestBranch: oldPipeline.DestBranch,
			Channels:   oldPipeline.Channels,
			PostHook:   oldPipeline.PostHook,
		})
	}

	for _, oldPostHook := range oldConf.PostHooks {
		newConf.PostHooks = append(newConf.PostHooks, config.PostHook(oldPostHook))
	}

	for _, oldDecoder := range oldConf.Decoders {
		newConf.Decoders = append(newConf.Decoders, config.Decoder(oldDecoder))
	}

	return &newConf, nil
}
