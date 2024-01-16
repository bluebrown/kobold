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

type options struct {
	config string
	confd  string
	dbfile string
	loglvl int
	logfmt string
	w      io.Writer
}

func (o *options) Bind(fs *flag.FlagSet) *options {
	if fs == nil {
		fs = flag.CommandLine
	}
	fs.StringVar(&o.config, "config", o.config, "path to config file")
	fs.StringVar(&o.confd, "confd", o.confd, "path to config dir")
	fs.StringVar(&o.dbfile, "db", o.dbfile, "path to sqlite db file")
	fs.IntVar(&o.loglvl, "loglvl", o.loglvl, "log level")
	fs.StringVar(&o.logfmt, "logfmt", o.logfmt, "log format, one of: json, text")
	return o
}

func Options() *options {
	var (
		dir = os.TempDir()
		err error
	)

	if dir, err = os.UserConfigDir(); err == nil {
		d := filepath.Join(dir, "kobold")
		if err = os.MkdirAll(d, 0755); err == nil {
			dir = d
		}
	}

	return &options{
		dbfile: filepath.Join(dir, "kobold.sqlite3"),
		loglvl: int(slog.LevelInfo),
		logfmt: "json",
		w:      os.Stderr,
	}
}
