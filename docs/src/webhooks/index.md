# Webhooks Events

Kobold listens for webhooks event using the
[endpoint](../configuration/endpoints.md) configurations from the config.yaml.
The webhook endpoints can have different types such as dockerhub or ACR. This is
because webhooks events from the various registries send different payloads.

Since kobold requires the digest to work, some registries require kobold to fetch the
digest upon receiving the event because its not part of the payload. If such a registry
is used, kobold requires to be able to authenticate against the registry.
