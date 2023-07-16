# Kobold

Kobold is a gitbot that automates the process of updating container image tags
in a git repository. It listens for webhook events triggered by a container
registry, and automatically updates the image tag in the git repository when a
new image is pushed.

## Motivation

Manually updating image tags in a git repository can be a time-consuming and
error-prone process. Kobold was created to automate this process and make it
easier for developers to keep their image tags up to date. It is meant to be a
companion to other gitops tools such as argocd or fluxcd, which will monitor for
changes in git repositories. So an image tag update by kobold might kick of a
application rollout with using the new version.

## Documentation

The documentation is hosted via github pages and can be viewed at
<https://bluebrown.github.io/kobold>.

## Quick Start

You can review the [example manifests](./manifests/example/) to see a possible
setup.

Typically you would deploy kobold to kubernetes providing your own config file
and environment variables for secrets.

You can create a kustomization using the dist overlay as base.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - https://github.com/bluebrown/kobold//manifests/dist/?rev=main
  - ingress.yaml
configMapGenerator:
  - name: kobold-config
    behavior: replace
    files:
      - ./etc/config.yaml
secretGenerator:
  - name: kobold-env
    behavior: replace
    envs:
      - ./etc/.env
```

The above config requires to add create some additional resources. Below is the
config file for kobold. Adjust this to your needs.

```yaml
endpoints:
  - name: myacr
    type: acr
    path: /acr/107ed4ca-591a-4fee-b6b2-65ef184bb582
    requiredHeaders:
      - key: Authorization
        value: ${SECRET_TOKEN}

repositories:
  - name: github-kobold
    url: https://github.com/bluebrown/kobold
    username: "${MY_GIT_EMAIL}"
    password: "${MY_GIT_PAT}"
    provider: github

subscriptions:
  - name: kobold
    endpointRefs:
      - name: myacr
    repositoryRef:
      name: github-kobold
    branch: main
```

In addition to the above config file, an env file is used for secrets. Kobold
will resolve env var references in the config file.

```console
MY_GIT_EMAIL=foo@bar.baz
MY_GIT_PAT=supersecret
SECRET_TOKEN=topsecret
```

Finally, you need an ingress if you intend to have an external registry dispatch
events to kobold. The below example matches the endpoint from the kobold config.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kobold
spec:
  rules:
    - http:
        paths:
          - pathType: Exact
            path: /acr/107ed4ca-591a-4fee-b6b2-65ef184bb582
            backend:
              service:
                name: kobold
                port:
                  name: http
```

## Development

Use the [Taskfile.yaml](./Taskfile.yaml) via `bin/task` to run various commands
useful for the development.
