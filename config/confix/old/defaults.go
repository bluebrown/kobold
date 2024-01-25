package old

import "os"

func (c *NormalizedConfig) Defaults() {
	if c.CommitMessage.Title == "" {
		c.CommitMessage.Title = "chore(kobold): update images"
	}

	if c.CommitMessage.Description == "" {
		c.CommitMessage.Description = `
{{- range . }}
- change {{ .Source }}[{{ .Parent }}]:
  - old: {{ .OldImageRef }}
  - new: {{ .NewImageRef }}
  - opt: {{ .OptionsExpression }}
{{- end }}
		`
	}

	// If the namespace is empty "", the underlying
	// auth package defaults to the "default" namespace.
	if c.RegistryAuth.Namespace == "" {
		c.RegistryAuth.Namespace = getEnv("NAMESPACE", "")
	}

	// If the service account is empty, set the special value
	// "no service account", which disabled service account lookup,
	// on the underlying package.
	if c.RegistryAuth.ServiceAccount == "" {
		c.RegistryAuth.ServiceAccount = getEnv("SERVICE_ACCOUNT_NAME", "no service account")
	}
}

func getEnv(key, value string) string {
	if s := os.Getenv(key); s != "" {
		value = s
	}
	return value
}
