# File Associations

In order to find image references, the file must be of a known type. Currently
known types are `kubernetes`, `docker-compose` and `ko-build`.

The type of a given yaml file is determined by using glob to math the filename.
The default matching rules look like this:

```yaml
- kind: ko-build
  pattern: ".ko.yaml"
- kind: docker-compose
  pattern: "*compose*.y?ml"
- kind: kubernetes
  pattern: "*"
```

In some cases it can be useful to overwrite the default rules. This can be done
per subscription.

```yaml
subscriptions:
  - name: example
    fileAssociations:
      - kind: docker-compose
        pattern: "dev.yaml"
      - kind: kubernetes
        pattern: "*"
```

> **Note** If there is a filetype you would to use, that is currently not
supported, please open and issue so we can  can add it to the codebase.
