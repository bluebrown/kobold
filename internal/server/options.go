package server

import "github.com/bluebrown/kobold/kobold/config"

func WithConfigPath(path string) Option {
	return func(o *Options) {
		o.ConfigPath = path
	}
}

func WithWatch(enabled bool) Option {
	return func(o *Options) {
		o.Watch = enabled
	}
}

func WithConfig(c *config.NormalizedConfig) Option {
	return func(o *Options) {
		o.Config = c
	}
}

func WithDataPath(path string) Option {
	return func(o *Options) {
		o.Datapath = path
	}
}

func WithK8sChain(o *Options) {
	o.UseK8sChain = true
}

func WithDefaultRegistry(registry string) Option {
	return func(o *Options) {
		o.defaultRegistry = registry
	}
}

func WithMuxGenerator(g muxGenerator) Option {
	return func(o *Options) {
		o.muxGenerator = g
	}
}

func WithImagerefTemplate(t string) Option {
	return func(o *Options) {
		o.imagerefTemplate = t
	}
}
