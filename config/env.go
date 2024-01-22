package config

import (
	"flag"
	"fmt"
	"strings"
)

// can be used to parse environment variables into flags with
// flag.VisitAll(UseEnv(os.Environ(), "SOME_PREFIX_")). the flag name is
// converted to uppercase and dashes are replaced with underscores. e.g.
// --foo-bar becomes FOO_BAR. The order of precedence is determined by the order
// of flag.VisitAll and flag.Parse calls.
func UseEnv(env []string, prefix string) func(*flag.Flag) {
	return func(f *flag.Flag) {
		name := strings.ReplaceAll(strings.ToUpper(f.Name), "-", "_")
		name = prefix + name

		f.Usage = fmt.Sprintf("%s (env: %s)", f.Usage, name)

		if val, ok := Lookup(env, name, ""); ok {
			if err := f.Value.Set(val); err != nil {
				panic(err)
			}
		}
	}
}

// searches env for key and returns the value if found, otherwise fallback is
// returned. Returns a boolen indicating if the key was found.
func Lookup(env []string, key, fallback string) (string, bool) {
	key += "="
	for _, e := range env {
		if len(e) > len(key) && e[:len(key)] == key {
			return e[len(key):], true
		}
	}
	return fallback, false
}
