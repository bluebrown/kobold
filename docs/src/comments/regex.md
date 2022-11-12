# Regular Expression

If you don't want to use exact matches and your team is not doing semver yet,
you can use the `regex` type to match any pattern.

```yaml
image: bluebrown/busybox # kobold: tag: sprint_\d+; type: regex
```
