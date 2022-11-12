# Pull Requests

By default kobold will commit directly to the configured branch of the
subscription. However, it can be configured to open pull requests if it detects
changes to a given subscription.

```yaml
subscriptions:
  - name: example
    strategy: pull-request
    branch: main
```

The branch name for the pull request will be `kobold/<epoch-time>` and the PR
will be made against branch in the subscription.

In order to use pull-requests, your git provider must be supported since git
itself has no concept of pull requests. Pull requests are made via rest api call
and hence provider specific.

Currently supported providers are `github` and `azure`.

The provider is inferred but can be configured via
[repository](./repositories.md).

> **Note** If your provider is not available, please open an issue so that we
can add it to the codebase.
