package kioutil

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
)

type GitPackageURI struct {
	Repo string `json:"repo,omitempty" toml:"repo"`
	Ref  string `json:"ref,omitempty" toml:"ref"`
	Pkg  string `json:"pkg,omitempty" toml:"pkg"`
}

// TODO: appending .git here can cause mismatching git-credentials
func (r *GitPackageURI) String() string {
	return fmt.Sprintf("%s.git@%s%s", r.Repo, r.Ref, r.Pkg)
}

var pattern = regexp.MustCompile(`^(?P<repo>.*)@(?P<ref>\w+)(?P<pkg>\/.+)?$`)

func (uri *GitPackageURI) MustUnmarshalText(s string) {
	if err := uri.UnmarshalText([]byte(s)); err != nil {
		panic(err)
	}
}

func (uri *GitPackageURI) UnmarshalText(b []byte) error {
	matches := pattern.FindStringSubmatch(string(b))
	if len(matches) == 0 {
		return fmt.Errorf("invalid git package uri: %q", string(b))
	}
	for i, name := range pattern.SubexpNames() {
		if i == 0 {
			continue
		}
		switch name {
		case "repo":
			uri.Repo = strings.TrimSuffix(matches[i], ".git")
		case "ref":
			uri.Ref = matches[i]
		case "pkg":
			uri.Pkg = strings.TrimSuffix(matches[i], "/")
		}
	}
	return nil
}

func (uri GitPackageURI) MarshalText() ([]byte, error) {
	return []byte(uri.String()), nil
}

func (uri *GitPackageURI) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		return fmt.Errorf("cannot convert %T to GitPackageURI", value)
	}
	return uri.UnmarshalText(b)
}

func (uri GitPackageURI) Value() (driver.Value, error) {
	if uri.Repo == "" {
		return nil, nil
	}
	return uri.String(), nil
}
