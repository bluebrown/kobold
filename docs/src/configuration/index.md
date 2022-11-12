# Configuration

Kobold parses a config file in order to wire up the webhook server. The config
file is composed of various sections that can be pieced together. This allows
for flexibility while working with multiple registries, repositories and
branches.

Kobold will detected changes to the configuration file and perform a graceful
reload, without interrupting currently running processes.

Environment variables in the config files are supported. They will be expected
according to linux conventions, meaning `$var` and `${var}` work.

Below is an example with all possible configurations. Many properties are
options. You can grab the
[json-schema](https://github.com/bluebrown/kobold/blob/main/hack/schema.json),
to learn more.

```yaml
version: v1
commitMessage:
  title: "chore(kobold): update images"
  description: |
    {{- range . }}
    - change {{ .Source }}[{{ .Parent }}]:
      - old: {{ .OldImageRef }}
      - new: {{ .NewImageRef }}
      - opt: {{ .OptionsExpression }}
    {{- end }}
registryAuth:
  namespace: ${NAMESPACE}
  serviceAccount: kobold
  imagePullSecrets:
    - name: regcreds
endpoints:
  - name: dockerhub
    type: dockerhub
    path: /dockerhub
    requiredHeaders: []
repositories:
  - name: kobold
    url: https://github.com/bluebrown/kobold
    username: "${GIT_USR}"
    password: "${GIT_PAT}"
    provider: github
subscriptions:
  - name: example
    endpointRefs:
      - name: dockerhub
    repositoryRef:
      name: kobold
    branch: main
    strategy: pull-request
    scopes:
      - env/dev/
      - env/stage/
    fileAssociations:
      - kind: ko-build
        pattern: ".ko.yaml"
      - kind: docker-compose
        pattern: "*compose*.y?ml"
      - kind: kubernetes
        pattern: "*"
```
