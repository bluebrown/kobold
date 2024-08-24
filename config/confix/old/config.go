package old

type Decoder struct {
	Name   string `toml:"name"`
	Script string `toml:"script"`
}

type Channel struct {
	Name    string `toml:"name"`
	Decoder string `toml:"decoder"`
}

type PostHook struct {
	Name   string `toml:"name"`
	Script string `toml:"script"`
}

type Pipeline struct {
	Name       string     `toml:"name"`
	RepoURI    PackageURI `toml:"repo_uri"`
	DestBranch string     `toml:"dest_branch"`
	Channels   []string   `toml:"channels"`
	PostHook   string     `toml:"post_hook"`
}

type Config struct {
	Version   string     `toml:"version"`
	Channels  []Channel  `toml:"channel"`
	Pipelines []Pipeline `toml:"pipeline"`
	PostHooks []PostHook `toml:"post_hook"`
	Decoders  []Decoder  `toml:"decoder"`
}
