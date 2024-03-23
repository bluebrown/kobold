package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bluebrown/kobold/config"
	"github.com/bluebrown/kobold/http/api"
	"github.com/bluebrown/kobold/http/webhook"
	"github.com/bluebrown/kobold/store"
	"github.com/bluebrown/kobold/store/schema"
	"github.com/bluebrown/kobold/task"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	_ "modernc.org/sqlite"
)

func init() {
	store.MustMakeUUID()
	store.MustMakeSha1()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, os.Args[1:], os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		cancel()
		time.Sleep(time.Second)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, env []string) error {
	var (
		set                      = flag.NewFlagSet("kobold-server", flag.ExitOnError)
		opts                     = config.NewOptions().Bind(set)
		handler     task.Handler = task.KoboldHandler
		webhookAddr              = ":8080"
		apiAddr                  = ":9090"
		maxprocs                 = 10
		debounce                 = 5 * time.Second
		prefix                   = ""
	)

	set.Var(&handler, "handler", "task handler, one of: print, kobold, error")
	set.StringVar(&webhookAddr, "addr-webhook", webhookAddr, "webhook listen address")
	set.StringVar(&apiAddr, "addr-api", apiAddr, "api listen address")
	set.IntVar(&maxprocs, "maxprocs", 10, "max number of concurrent runs")
	set.DurationVar(&debounce, "debounce", time.Minute, "debounce interval for webhook events")
	set.StringVar(&prefix, "prefix", prefix, "prefix for all routes, must NOT contain trailing slash")

	set.VisitAll(config.UseEnv(env, "KOBOLD_"))

	if err := set.Parse(args); err != nil {
		return fmt.Errorf("parse args: %w", err)
	}

	query, err := config.Configure(ctx, *opts, schema.TaskSchema, schema.ReadSchema)
	if err != nil {
		return fmt.Errorf("configure: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	sched := task.NewScheduler(ctx, query, maxprocs)

	g.Go(func() error {
		sched.SetHandler(handler)
		return sched.Run(debounce)
	})

	g.Go(func() error {
		apmux := http.NewServeMux()
		apmux.Handle(prefix+"/api/", http.StripPrefix(prefix+"/api", api.New(prefix+"/api", query)))
		apmux.Handle(prefix+"/metrics", promhttp.Handler())
		return listenAndServeContext(ctx, "api", apiAddr, apmux)
	})

	g.Go(func() error {
		whmux := http.NewServeMux()
		whmux.Handle(prefix+"/", http.StripPrefix(prefix, webhook.New(sched)))
		return listenAndServeContext(ctx, "webhook", webhookAddr, whmux)
	})

	return g.Wait()
}

func listenAndServeContext(ctx context.Context, name, addr string, handler http.Handler) error {
	slog.InfoContext(ctx, "server startup", "name", name, "addr", addr)

	if err := ctx.Err(); err != nil {
		return err
	}

	var (
		server = http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second}
		errc   = make(chan error, 1)
	)

	go func() {
		defer close(errc)
		errc <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.InfoContext(ctx, "server shutdown", "name", name)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(ctx) //nolint:contextcheck
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("server %s: %w", name, err)
	}
}
