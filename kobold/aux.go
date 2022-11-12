package kobold

import "strings"

// infer the provider from the given url. If the provider
// cannot be inferred, an empty string is returned
//
// TODO: handle edge cases. i.e. azure repo url contains
// github.com for some reason
func InferGitProvider(giturl string) GitProvider {
	if strings.Contains(giturl, "github.com") {
		return ProviderGithub
	}
	if strings.Contains(giturl, "dev.azure.com") {
		return ProviderAzure
	}
	return ""
}
