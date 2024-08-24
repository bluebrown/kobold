package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/bluebrown/kobold/config/confix"
	"github.com/bluebrown/kobold/config/confix/old"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		target    = "config.yaml"
		writeback = false
	)

	flag.StringVar(&target, "f", target, "target file")
	flag.BoolVar(&writeback, "w", writeback, "write back to target file")
	flag.Parse()

	oc := old.Config{}
	if err := old.ReadFile(target, &oc); err != nil {
		return fmt.Errorf("read: %w", err)
	}

	v2, err := confix.MakeConfig(&oc)
	if err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(v2); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	if !writeback {
		fmt.Println(buf.String())
		return nil
	}

	if err := os.WriteFile(target, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}
