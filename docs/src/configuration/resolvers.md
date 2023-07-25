# Resolvers

It is possible to create custom resolvers, to look up image references in yaml
files.

```yaml
resolvers:
  - name: my-custom
    paths:
      - path.to.image
      - another.path
```

If a resolvers is used, via a [file association](./file-associations.md), image
references at all paths of the resolver are handled. Kobold does not stop on
first match. Additionally, if a path does not exist, kobold will continue with
the next path, without error or warning.
