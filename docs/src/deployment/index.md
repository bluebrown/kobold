# Deployment

The kobold binary accepts so optional flags. These can be used to influence the runtime behavior.

```toml
Usage of kobold:
  -config string
        path to the config file (default "path/to/kobold/config.yaml")
  -data string
        path to temporary data (default "/tmp/kobold")
  -default-registry string
        the default registry to use, for unprefixed images (default "index.docker.io")
  -imageref-template string
        the format of the image ref when updating an image node (default "{{ .Image }}:{{ .Tag }}@{{ .Digest }}")
  -k8schain
        use k8schain for registry authentication
  -log-format string
        the log format, console or json (default "json")
  -port int
        set the server port (default 8080)
  -v int
        verbosity level. 0 is fatal - 7 is trace (default 5)
  -version
        show version info
  -watch
        Reload the server on config file change
  -debounce duration
        debounce events until no event has been received for the provided duration (default 1m0s)
```
