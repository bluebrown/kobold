package config

import (
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

var UsePragmas = []string{
	"busy_timeout=5000",
	"journal_mode=WAL",
	"foreign_keys=on",
}

type Options struct {
	Config string
	Confd  string
	dbfile string
	Loglvl int
	Logfmt string
	W      io.Writer
}

func (o *Options) Bind(fs *flag.FlagSet) *Options {
	if fs == nil {
		fs = flag.CommandLine
	}
	fs.StringVar(&o.Config, "config", o.Config, "path to config file")
	fs.StringVar(&o.Confd, "confd", o.Confd, "path to config dir")
	fs.StringVar(&o.dbfile, "db", o.dbfile, "path to sqlite db file")
	fs.IntVar(&o.Loglvl, "loglvl", o.Loglvl, "log level")
	fs.StringVar(&o.Logfmt, "logfmt", o.Logfmt, "log format, one of: json, text")
	return o
}

func NewOptions() *Options {
	dir := os.TempDir()

	if cd, err := os.UserConfigDir(); err == nil {
		d := filepath.Join(cd, "kobold")
		if err = os.MkdirAll(d, 0o755); err == nil {
			dir = d
		}
	}

	return &Options{
		dbfile: filepath.Join(dir, "kobold.sqlite3"),
		Loglvl: int(slog.LevelInfo),
		Logfmt: "json",
		W:      os.Stderr,
	}
}
