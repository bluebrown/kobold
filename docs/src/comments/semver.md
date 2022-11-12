# Semantic Versioning

It is best practice to use semantic versioning. With the type `semver`, you can
match image tags that adhere to the semver spec.

```yaml
image: bluebrown/busybox # kobold: tag: ^1; type: semver
```

The incoming tag is matched with the tag in the comment option using
[Masterminds semver](https://github.com/Masterminds/semver). You can review the
[comparison rules](https://github.com/Masterminds/semver#basic-comparisons) in
their readme to learn more.
