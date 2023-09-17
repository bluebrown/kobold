# Registry Auth

Depending on the image registry, the push event payload may not contain all
required data. In such cases, kobold will fetch the missing data from the given
registry. Because of this it may need to authenticate against the registry.

The [authn
package](https://github.com/google/go-containerregistry/tree/main/pkg/authn) is
used for authentication. You can review their repository to learn more. It will
be either the default or k8s chain.

## Kubernetes

When running in kubernetes, the registryAuth property can be used to access
image pull secrets. For this the `--k8schain` switch must be set.
ImagePullSecrets are taken from the service account in the configuration and
from imagePullSecrets listed. Both service account and imagePullSecrets are
looked up in the given namespace.

```yaml
registryAuth:
  namespace: ${NAMESPACE}
  serviceAccount: kobold
  imagePullSecrets:
    - name: regcreds
```

When enabling the k8schain, you *may* need to create rbac resources allowing
kobold to lookup services accounts and secrets.

The below rbac resources config will be created if you use the provided
kustomization, to ensure kobold is allowed to lookup service accounts and
secrets in its own namespace.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kobold
automountServiceAccountToken: true
imagePullSecrets: []
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: default
  name: k8schain
rules:
  - apiGroups: [""]
    resources: ["serviceaccounts", "secrets"]
    verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: kobold-k8schain
  namespace: default
subjects:
  - kind: ServiceAccount
    name: kobold
    namespace: default
roleRef:
  kind: Role
  name: k8schain
  apiGroup: rbac.authorization.k8s.io
```

You can disable the lookup of the service account by using the value `no service
account` as service account name.

```yaml
registryAuth:
  serviceAccount: "no service account"
```

By doing this, you could remove the resource `serviceaccounts` from the rbac
role.

If you don't provide a `registryAuth` key in your config and use the k8s chain,
kobold will default to the values of the environment variables `$NAMESPACE` and
`$SERVICE_ACCOUNT_NAME`, if set. Otherwise it will use the `default` namespace and
`no service account`. This allows to use the other parts of the k8s chain
without the actual need for rbac.

## Bare Metal

If the flag is not used, kobold will attempt to use your local configuration to
authenticate. If you have logged in via docker cli, for example, kobold will be
able to access your private repository.
