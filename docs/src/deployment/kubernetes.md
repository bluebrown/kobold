# Kubernetes

You can use the provided kustomization as resource and overwrite the config and
env secret with your own.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kobold
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

This will also create rbac rules and use the k8schain as shown in
[registry-auth](../configuration/registry-auth.md).

> **Note** Your config file should contain at least the version. Currently only
> `v1` exists.

## Ingress

Since kobold listens for webhook event, you probably want to deploy an ingress.
It is recommended to use UUIDS for your path so that the path is not guessable.
This is especially important of your registry does not support custom headers.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kobold
  labels:
    app.kubernetes.io/name: kobold
spec:
  rules:
    - http:
        paths:
          - pathType: Exact
            path: "/dockerhub/9a06938d-4022-46d3-8528-82cb95ee1ad5"
            backend:
              service:
                name: kobold
                port:
                  name: http
```
