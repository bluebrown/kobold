package config

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/bluebrown/kobold/store/model"
	ksql "github.com/bluebrown/kobold/store/schema"
)

func Configure(ctx context.Context, opts Options, schemas ...[]byte) (*model.Queries, error) {
	SetLog(opts.W, opts.Logfmt, slog.Level(opts.Loglvl))

	sqliteDSN := "file:" + opts.dbfile + "?" + query(UsePragmas)

	db, err := sql.Open("sqlite", sqliteDSN)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db %q: %w", sqliteDSN, err)
	}

	if opts.Config != "" || opts.Confd != "" {
		slog.InfoContext(ctx, "db config purge", "reason", "config file(s) provided")
		schemas = append(schemas, ksql.CleanConfig)
	}

	for _, schema := range schemas {
		if _, err := db.ExecContext(ctx, string(schema)); err != nil {
			return nil, fmt.Errorf("create schema: %w", err)
		}
	}

	model := model.New(db)

	if err := ApplyBuiltins(ctx, model); err != nil {
		return nil, fmt.Errorf("apply builtins: %w", err)
	}

	var cfg Config

	if opts.Config != "" {
		if err := ReadFile(opts.Config, &cfg); err != nil {
			return nil, fmt.Errorf("read %s: %w", opts.Config, err)
		}
	}

	if opts.Confd != "" {
		if err := ReadConfD(opts.Confd, &cfg); err != nil {
			return nil, fmt.Errorf("read %s: %w", opts.Confd, err)
		}
	}

	if err := cfg.Apply(ctx, model); err != nil {
		return nil, fmt.Errorf("apply %q to db: %w", opts.Config, err)
	}

	return model, nil
}

func query(prags []string) string {
	return (url.Values{"_pragma": prags}).Encode()
}
