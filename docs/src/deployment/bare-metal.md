# Bare Metal

You can run kobold via binary from the [release
page](https://github.com/bluebrown/kobold/releases).

## Binary

```bash
./kobold
```

## Docker

If you want to run kobold via docker, mount your config file to /ko-app/config.yaml
or use the `--config` flag.

```bash
docker run -v "$PWD/config.yaml:/ko-app/config.yaml" -p 8080:8080 bluebrown/kobold
```
