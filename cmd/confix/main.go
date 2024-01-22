package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
		target = "config.yaml"
		out    = ""
	)

	flag.StringVar(&target, "f", target, "target file")
	flag.StringVar(&out, "o", out, "output directory")
	flag.Parse()

	if out != "" {
		if err := os.MkdirAll(out, 0755); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
	}

	v1, err := old.ReadPath(target)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	v1.Defaults()

	v2, err := confix.MakeConfig(v1)
	if err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(v2); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	gc, err := confix.MakeGitCredentials(v1)
	if err != nil {
		return fmt.Errorf("git-credentials: %w", err)
	}

	if err := os.WriteFile(filepath.Join(out, "kobold.toml"), buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	if len(gc) < 1 {
		return nil
	}

	if err := os.WriteFile(filepath.Join(out, ".git-credentials"), []byte(gc), 0600); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	if err := os.WriteFile(filepath.Join(out, ".gitconfig"), confix.MakeGitConfig(), 0600); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}
