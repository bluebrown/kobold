package config

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"

	ksql "github.com/bluebrown/kobold/sql"
	"github.com/bluebrown/kobold/store"
)

func Configure(ctx context.Context, opts options, schemas ...[]byte) (*store.Queries, error) {
	SetLog(opts.w, opts.logfmt, slog.Level(opts.loglvl))

	sqliteDSN := "file:" + opts.dbfile + "?" + query(UsePragmas)

	db, err := sql.Open("sqlite", sqliteDSN)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db %q: %w", sqliteDSN, err)
	}

	if opts.config != "" || opts.confd != "" {
		slog.InfoContext(ctx, "db config purge", "reason", "config file(s) provided")
		schemas = append(schemas, ksql.CleanConfig)
	}

	for _, schema := range schemas {
		if _, err := db.ExecContext(ctx, string(schema)); err != nil {
			return nil, fmt.Errorf("create schema: %w", err)
		}
	}

	model := store.New(db)

	if err := ApplyBuiltins(ctx, model); err != nil {
		return nil, fmt.Errorf("apply builtins: %w", err)
	}

	var cfg Config

	if opts.config != "" {
		if err := ReadFile(opts.config, &cfg); err != nil {
			return nil, fmt.Errorf("read %s: %w", opts.config, err)
		}
	}

	if opts.confd != "" {
		if err := ReadConfD(opts.confd, &cfg); err != nil {
			return nil, fmt.Errorf("read %s: %w", opts.confd, err)
		}
	}

	if err := cfg.Apply(ctx, model); err != nil {
		return nil, fmt.Errorf("apply %q to db: %w", opts.config, err)
	}

	return model, nil
}

func query(prags []string) string {
	return (url.Values{"_pragma": prags}).Encode()
}
