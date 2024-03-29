package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/bluebrown/kobold/config"
	"github.com/bluebrown/kobold/store"
	"github.com/bluebrown/kobold/store/schema"
	"github.com/bluebrown/kobold/task"
	_ "modernc.org/sqlite"
)

func init() {
	store.MustMakeUUID()
	store.MustMakeSha1()
}

func main() {
	var input io.Reader

	if info, err := os.Stdin.Stat(); err == nil {
		if info.Mode()&os.ModeCharDevice == 0 {
			input = os.Stdin
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, os.Args[1:], os.Environ(), input); err != nil {
		fmt.Fprintln(os.Stderr, err)
		cancel()
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, env []string, input io.Reader) error {
	var (
		channel  string
		handler  task.Handler = task.KoboldHandler
		set                   = flag.NewFlagSet("kobold-cli", flag.ExitOnError)
		opts                  = config.NewOptions().Bind(set)
		maxprocs              = 10
	)

	set.StringVar(&channel, "channel", "", "channel to publish msgs to")
	set.Var(&handler, "handler", "task handler, must be one of: print, kobold, error")
	set.IntVar(&maxprocs, "maxprocs", 10, "max number of concurrent runs")

	set.VisitAll(config.UseEnv(env, "KOBOLD_"))

	if err := set.Parse(args); err != nil {
		return fmt.Errorf("parse args: %w", err)
	}

	query, err := config.Configure(ctx, *opts, schema.TaskSchema, schema.ReadSchema)
	if err != nil {
		return fmt.Errorf("configure: %w", err)
	}

	pool := task.NewPool(ctx, maxprocs, query)
	pool.SetHandler(handler)

	if input != nil {
		if err := pool.QueueReader(ctx, channel, input); err != nil {
			return fmt.Errorf("queue input: %w", err)
		}
	}

	if err := pool.Dispatch(); err != nil {
		return fmt.Errorf("dispatch: %w", err)
	}

	if err := pool.Wait(); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	return nil
}
