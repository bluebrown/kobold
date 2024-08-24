package old

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/a8m/envsubst"
)

func ReadFile(path string, cfg *Config) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("get abs path for %q: %w", path, err)
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("eval symlink %q: %w", path, err)
	}

	b, err := envsubst.ReadFile(path)
	if err != nil {
		return fmt.Errorf("envsubst read file %q: %w", path, err)
	}

	err = toml.Unmarshal(b, cfg)
	if err != nil {
		return fmt.Errorf("unmarshal toml: %w", err)
	}

	return nil
}

func ReadConfD(dir string, cfg *Config) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir %q: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		if err := ReadFile(path, cfg); err != nil {
			return fmt.Errorf("read file %q: %w", path, err)
		}
	}

	return err
}
