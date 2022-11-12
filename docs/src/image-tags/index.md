# Image Tags

By default, the image tags in the git repository are updated with the digest of
the container image, rather than just the tag.

The digest is a unique identifier that is generated based on the contents of the
image, so even if the tag remains the same, the digest will change if the
contents of the image are updated. This allows kobold to stay up to date with
the correct image, even if the same tag is pushed multiple times.

The behavior can be configured using the `--imageref-template` flag. The default
value is the below.

```console
{{ .Image }}:{{ .Tag }}@{{ .Digest }}
```

This string must be a valid go-template. The data passed into this template is
the [`PushEvent`](./internal/events/events.go).

Note that that when using the default form, the tag is only there for
informational purposes. This is because most registries will [ignore the tag if
a digest is
present](https://github.com/distribution/distribution/blob/362910506bc213e9bfc3e3e8999e0cfc757d34ba/reference/normalize.go#L88).
