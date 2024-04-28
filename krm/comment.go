package krm

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
)

const (
	TypeExact  = "exact"
	TypeSemver = "semver"
	TypeRegex  = "regex"
)

const (
	KeyType    = "type"
	KeyTag     = "tag"
	KeyPart    = "part"
	KeyContext = "context"
)

type Options struct {
	Type    string
	Tag     string
	Part    string
	Context string
}

func ParseOpts(expr string) (Options, error) {
	kvs := strings.Split(strings.TrimSuffix(expr, ";"), ";")
	opts := Options{}
	for _, kv := range kvs {
		k, v, ok := strings.Cut(kv, ":")
		if !ok {
			return opts, fmt.Errorf("invalid key value pair: %s", kv)
		}
		switch strings.TrimSpace(k) {
		case KeyType:
			opts.Type = strings.TrimSpace(v)
		case KeyTag:
			opts.Tag = strings.TrimSpace(v)
		case KeyPart:
			opts.Part = strings.TrimSpace(v)
		case KeyContext:
			opts.Context = strings.TrimSpace(v)
		default:
			return opts, fmt.Errorf("unknown key: %s", v)
		}
	}
	return opts, nil
}

// check if the provided tag matches per options,
// Exact, semver or regex.
func MatchTag(tag string, opts Options) (bool, error) {
	switch opts.Type {
	case TypeExact:
		if tag != opts.Tag {
			return false, nil
		}
	case TypeSemver:
		c, err := semver.NewConstraint(opts.Tag)
		if err != nil {
			return false, fmt.Errorf("could not parse version constraint from opt: %w", err)
		}
		v, err := semver.NewVersion(tag)
		if err != nil || !c.Check(v) {
			return false, nil //nolint:nilerr
		}
	case TypeRegex:
		ok, err := regexp.MatchString(fmt.Sprintf("^%s$", opts.Tag), tag)
		if err != nil {
			return false, fmt.Errorf("invalid regex in opt: %w", err)
		}
		if !ok {
			return false, nil
		}
	default:
		return false, fmt.Errorf("type %q is not supported", opts.Type)
	}
	return true, nil
}
