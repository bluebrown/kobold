# Repositories

The repositories section of the config file specifies the git repositories that
Kobold should update when it receives a webhook event. The presence of a
repository does not lead to any actual updates when an event is received. The
repository functions as abstract type that can be referenced by subscriptions.

```yaml
repositories:
- name: kobold
    url: https://github.com/bluebrown/kobold
    username: "${GIT_USR}"
    password: "${GIT_PAT}"
    provider: github
```

At the moment, only password/token authentication is supported. This is because
for [pull-requests](./pull-requests.md), you need to use password authentication
regardless. So it is more easy to share the authentication config for regular
git commands and [pull-requests](./pull-requests.md).

> **Note** that the provider is not required but may help if kobold is not able
to infer the provider based on the url.
