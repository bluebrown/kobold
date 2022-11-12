package kobold

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"sigs.k8s.io/yaml"
)

type normalizer interface {
	Normalize() *NormalizedConfig
}

// read the config at the given path and expand its env var references the
// config will be coded into an intermediate struct based on the version and
// afterwards normalized
func ReadPath(path string) (*NormalizedConfig, error) {
	b, err := readExpandFile(path)
	if err != nil {
		return nil, err
	}

	m := map[string]any{}
	yaml.Unmarshal(b, &m)

	var norm normalizer
	switch v := m["version"]; {
	case v == "v1":
		norm = &UserConfigV1{}
	case v == nil:
		return nil, fmt.Errorf("config version is required")
	default:
		return nil, fmt.Errorf("unsupported config version %q", v)
	}

	err = mapstructure.Decode(m, norm)
	if err != nil {
		return nil, err
	}

	return norm.Normalize(), err
}

func readExpandFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return b, err
	}
	return []byte(os.ExpandEnv(string(b))), nil
}
