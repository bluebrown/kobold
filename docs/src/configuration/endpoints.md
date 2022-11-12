# Endpoints

The endpoints section of the config file specifies the webhook endpoints that
Kobold should listen to. These can be of different types, such as Azure
Container Registry or Docker Hub.

Each endpoint must have a unique name, a type, a path and a optional list of
requiredHeaders. The path is the URL path that the endpoint listens to, and the
requiredHeaders are HTTP headers that must be included in the webhook request
for it to be accepted by Kobold.

```yaml
endpoints:
  - name: myacr
    type: acr
    path: /acr/292d91a8-d073-4a65-99b8-0018fa6f8f46
   requiredHeaders:
      - key: Authorization
        value: "Basic ${BASE64AUTH}"
```

## Types

Currently supported endpoint types are `acr`, `dockerhub` and `generic`. The
generic type can be used if you want to dispatch events manually, perhaps via
pipeline.

> **Note** If there is no type for your registry of choice, please open an issue so that we
can add it to the codebase.
