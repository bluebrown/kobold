package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/rs/zerolog/log"

	"github.com/bluebrown/kobold/internal/gitbot/transport"
	"github.com/bluebrown/kobold/internal/logging"
	"github.com/bluebrown/kobold/internal/server"
)

var (
	version          = "unknown"
	showVersion      = false
	configPath       string
	dataPath         string
	port             = 8080
	useK8sChain      = false
	defaultRegistry  = name.DefaultRegistry
	imageRefTemplate = "{{ .Image }}:{{ .Tag }}@{{ .Digest }}"
	watch            = false
)

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	flag.BoolVar(&showVersion, "version", showVersion, "show version info")
	flag.IntVar(&port, "port", port, "set the server port")
	flag.StringVar(&configPath, "config", filepath.Join(dir, "config.yaml"), "path to the config file")
	flag.StringVar(&dataPath, "data", filepath.Join(os.TempDir(), "kobold"), "path to temporary data")
	flag.BoolVar(&useK8sChain, "k8schain", useK8sChain, "use k8schain for registry authentication")
	flag.StringVar(&defaultRegistry, "default-registry", defaultRegistry, "the default registry to use, for unprefixed images")
	flag.StringVar(&imageRefTemplate, "imageref-template", imageRefTemplate, "the format of the image ref when updating an image node")
	flag.BoolVar(&watch, "watch", watch, "Reload the server on config file change")
	logging.InitFlags(nil)
	flag.Parse()

	if showVersion {
		fmt.Printf("Kobold Version: %s\n", version)
		fmt.Printf("Kobold Git:     %s\n", transport.Tag)
		os.Exit(0)
	}

	logging.ConfigureLogging()

	if err := run(); err != nil {
		log.Error().Err(err).Msg("error while running server")
	}
}

func run() error {
	if err := os.RemoveAll(dataPath); err != nil {
		return err
	}

	m := http.NewServeMux()

	opts := []server.Option{
		server.WithImagerefTemplate(imageRefTemplate),
		server.WithDefaultRegistry(defaultRegistry),
		server.WithDataPath(dataPath),
		server.WithConfigPath(configPath),
		server.WithWatch(watch),
	}

	if useK8sChain {
		opts = append(opts, server.WithK8sChain)
	}

	m.Handle("/", server.NewOrDie(opts...))

	// implement commonly used health check endpoints
	m.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {})
	m.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {})
	m.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {})

	log.Info().Int("port", port).Msg("starting server")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), m); err != http.ErrServerClosed {
		return err
	}

	return nil
}
