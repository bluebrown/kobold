# Scopes

Each subscription can be scoped to one or more path. The path in the scope list
are inverted gitignore rules. So the matching works the same as in a .gitignore.

```yaml
subscriptions:
  - name: example
    scopes:
      - /env/prod/
      - /docker-compose.yaml
```

This is useful if you want to use different strategies within the same
repository. For example you could use the `commit` strategy for your staging
environment and `pull-request` for the production environment.  

It could be used to rollout new application versions automatically to a staging
environment, for review by your stakeholders. Once they are happy, you can merge
the pull request for your production environment.

```yaml
subscriptions:
  - name: stage
    strategy: commit
    scopes:
      - /env/stage/

  - name: prod
    strategy: pull-request
    scopes:
      - /env/prod/
```
