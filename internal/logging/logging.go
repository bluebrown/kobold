package logging

import (
	"flag"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// the log format
	logFormat = "json"
	// the log level
	logLevel = 5
)

// configure logging. Should be called after flag.Parse
func ConfigureLogging() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.TimestampFieldName = "ts"
	zerolog.LevelFieldName = "lvl"
	zerolog.MessageFieldName = "msg"

	if strings.ToLower(logFormat) == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.SetGlobalLevel(zerolog.Level(6 - logLevel))
	zerolog.DefaultContextLogger = &log.Logger
}

// initialize the flags. should be called before flag.Parse
func InitFlags(flagset *flag.FlagSet) {
	if flagset == nil {
		flagset = flag.CommandLine
	}
	flagset.StringVar(&logFormat, "log-format", logFormat, "the log format, console or json")
	flagset.IntVar(&logLevel, "v", logLevel, "verbosity level. 0 is fatal - 7 is trace")
}
