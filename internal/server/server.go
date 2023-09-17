package server

import (
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"

	"github.com/bluebrown/kobold/kobold/config"
)

type muxGenerator interface {
	Generate(conf *config.NormalizedConfig) (http.Handler, error)
}

type Server struct {
	atomicHandler *atomic.Value
	generator     muxGenerator
}

type Options struct {
	Watch            bool
	ConfigPath       string
	Config           *config.NormalizedConfig
	Datapath         string
	UseK8sChain      bool
	muxGenerator     muxGenerator
	defaultRegistry  string
	imagerefTemplate string
	debounce         time.Duration
}

type Option func(*Options)

func NewOrDie(options ...Option) *Server {
	opts := &Options{}

	for _, o := range options {
		o(opts)
	}

	if opts.ConfigPath != "" && opts.Config != nil {
		panic("configPath and config are mutually exclusive")
	}

	if opts.ConfigPath != "" {
		var err error
		opts.Config, err = config.ReadPath(opts.ConfigPath)
		if err != nil {
			panic(err)
		}
	}

	if opts.Config == nil {
		opts.Config = &config.NormalizedConfig{}
		opts.Config.Defaults()
	}

	if opts.Datapath == "" {
		opts.Datapath = filepath.Join(os.TempDir(), "kobold")
	}

	if opts.muxGenerator == nil {
		opts.muxGenerator = generator{
			dataDir:          opts.Datapath,
			useK8sChain:      opts.UseK8sChain,
			defaultRegistry:  opts.defaultRegistry,
			imagerefTemplate: opts.imagerefTemplate,
			debounceTime:     opts.debounce,
		}
	}

	s := &Server{
		generator:     opts.muxGenerator,
		atomicHandler: &atomic.Value{},
	}

	mux, err := s.generator.Generate(opts.Config)
	if err != nil {
		panic(err)
	}

	s.atomicHandler.Store(mux)

	if opts.Watch && opts.ConfigPath != "" {
		go WatchConfigOrDie(opts.ConfigPath, func(c *config.NormalizedConfig) {
			log.Info().Msg("reloading config")
			m, err := s.generator.Generate(c)
			if err != nil {
				log.Error().Err(err).Msg("could not reload config")
				return
			}
			s.Reload(m)
		})
	}

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.atomicHandler.Load().(http.Handler).ServeHTTP(w, r)
}

func (s *Server) Reload(handler http.Handler) {
	s.atomicHandler.Store(handler)
}

func WatchConfigOrDie(path string, onChange func(c *config.NormalizedConfig)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	err = watcher.Add(path)
	if err != nil {
		panic(err)
	}
	log.Debug().Str("path", path).Msg("watching config")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			log.Trace().Str("op", event.Op.String()).Msg("inotify event received")

			// if its not an event that modifies the file, ignore it
			if !(event.Op.Has(fsnotify.Write) || event.Op.Has(fsnotify.Create) || event.Op.Has(fsnotify.Remove)) {
				continue
			}

			// if the file has been removed, rewatch it. This helps with
			// scenarios where the file is deleted and moved or symlinked
			// to get atomic writes. I.e. in kubernetes this is the case
			if event.Op.Has(fsnotify.Remove) {
				if err := watcher.Add(event.Name); err != nil {
					log.Error().Err(err).Msg("failed to re-watch config")
					continue
				}
			}

			// finally load the new config

			conf, err := config.ReadPath(path)
			if err != nil {
				log.Error().Err(err).Msg("failed to config")
				continue
			}

			// call the on change handler
			onChange(conf)

		case err, ok := <-watcher.Errors:
			log.Error().Err(err).Msg("error while watching config file")
			if !ok {
				return
			}
		}
	}
}
