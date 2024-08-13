package git

import (
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
)

// the package uri is a special uri format to specify a git repo, ref, and
// package path within the repo under the ref. The uri is of the form:
//
//	<repo>?=ref=<ref>&pkg=<pkg>]
//
// where <repo> is the git repo uri, <ref> is the git ref (branch, tag, commit),
// and <pkg> is the package path within the repo. If <pkg> is not specified, the
// root of the repo is assumed.
type PackageURI struct {
	Repo string `json:"repo,omitempty" toml:"repo"`
	Ref  string `json:"ref,omitempty"  toml:"ref"`
	Pkg  string `json:"pkg,omitempty"  toml:"pkg"`
}

// FIXME: appending .git can lead to confusion or invalid URIs.
func (uri *PackageURI) String() string {
	return fmt.Sprintf("%s.git@%s%s", uri.Repo, uri.Ref, uri.Pkg)
}

func (uri *PackageURI) MustUnmarshalText(s string) {
	if err := uri.UnmarshalText([]byte(s)); err != nil {
		panic(err)
	}
}

func (uri *PackageURI) UnmarshalText(b []byte) error {
	var packageURIString = string(b)
	if !strings.Contains(packageURIString, "?") {
		return fmt.Errorf("invalid git package uri: %q, query params are missing", string(b))
	}
	var repoParts = strings.Split(packageURIString, "?")
	var repo = repoParts[0]
	var queryParams, err = url.ParseQuery(repoParts[1])
	if err != nil {
		return fmt.Errorf("invalid git package uri: %q, could not parse query params", string(b))
	}
	var ref = queryParams.Get("ref")
	if ref == "" {
		return fmt.Errorf("invalid git package uri: %q, missing ref query param", string(b))
	}
	var pkg = queryParams.Get("pkg")
	uri.Repo = strings.TrimSuffix(repo, ".git")
	uri.Ref = ref
	uri.Pkg = pkg
	return nil
}

func (uri PackageURI) MarshalText() ([]byte, error) {
	return []byte(uri.String()), nil
}

func (uri *PackageURI) Scan(value interface{}) error {
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
		return fmt.Errorf("cannot convert %T to PackageURI", value)
	}
	return uri.UnmarshalText(b)
}

func (uri PackageURI) Value() (driver.Value, error) {
	if uri.Repo == "" {
		return nil, nil
	}
	return uri.String(), nil
}
