# Comments

Once a webhook event is received, kobold performs a search in any
[repository](../configuration/repositories.md) that is
[subscribed](../configuration/subscriptions.md) to the origin
[endpoint](../configuration/endpoints.md). It looks for image nodes containing
an inline comment with some options to configure the behavior on case by case
basis.

Inline comments keep the logic lean and avoid verbosity while preserving valid
yaml.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - name: my-app
    image: org/app # kobold: tag: ^1; type: semver
```

## Format

The comments parsed by kobold have the following format:

```console
# kobold: tag: <tag-name>; type: <tag-type>
```

The tag-type can be either [exact](./exact.md), [semver](./semver.md), or
[regex](regex.md), and specifies how kobold should interpret the tag-name.

For example, if tag-type is semver, the tag-name can include common semantic
versioning semantics, such as ^1 to denote that any tag between v1 and v2 should
be matched (not including v2).
