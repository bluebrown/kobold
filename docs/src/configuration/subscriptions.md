# Subscriptions

The subscriptions section of the config file specifies the rules that determine
which webhook events should trigger updates to which repositories. Each
subscription must have a name, a list of endpointRefs, a repositoryRef, a
branch, a strategy, and a list of scopes.

The `endpointRefs` are references to the [endpoints](./endpoints.md) specified
in the endpoints section that should trigger this subscription.

The `repositoryRef` is a reference to the [repository](./repositories.md)
specified in the repositories section that should be updated, the branch is the
branch of the repository to update.

The `strategy` is the update strategy and can be either commit or
[pull-request](./pull-requests.md).

The [`scopes`](./scopes.md) are the file or directory paths within the
repository to operate on. This allows to subscribe with a given repository
multiple times using different branches, scopes and strategies.

```yaml
subscriptions:
  - name: example
    endpointRefs:
      - name: my-dockerhub
      - name: my-acr
    repositoryRef:
      name: kobold
    branch: main
    strategy: pull-request
    scopes:
      - env/prod/
```
