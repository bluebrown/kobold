# File Associations

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

If the builtin resolvers are not sufficient. You can [create your own
resolver](./resolvers.md).
