package registry

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/authn/kubernetes"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/rs/zerolog/log"

	"github.com/bluebrown/kobold/kobold/config"
)

func NewKeys(k8s bool, auth config.RegistryAuthSpec) (keys authn.Keychain, err error) {
	if k8s {
		log.Debug().Msg("using k8s key chain")
		ss := make([]string, len(auth.ImagePullSecrets))
		for i, s := range auth.ImagePullSecrets {
			ss[i] = s.Name
		}
		keys, err = k8schain.NewInCluster(context.TODO(), kubernetes.Options{
			Namespace:          auth.Namespace,
			ServiceAccountName: auth.ServiceAccount,
			ImagePullSecrets:   ss,
		})
		if err != nil {
			return nil, err
		}
	} else {
		log.Debug().Msg("using default key chain")
		keys = authn.DefaultKeychain
	}
	return keys, nil
}

func fetchDigest(ref, defaultRegistry string, keys authn.Keychain) (digest v1.Hash, err error) {
	n, err := name.ParseReference(ref, name.WithDefaultRegistry(defaultRegistry))
	if err != nil {
		return digest, err
	}
	r, err := remote.Get(n, remote.WithAuthFromKeychain(keys))
	if err != nil {
		return digest, err
	}
	return r.Digest, nil
}

type DigestFetcher struct {
	defaultRegistry string
	keys            authn.Keychain
}

func (df DigestFetcher) Fetch(ref string) (string, error) {
	d, err := fetchDigest(ref, df.defaultRegistry, df.keys)
	if err != nil {
		return "", err
	}
	return d.String(), nil
}

func NewDigestFetcher(defaultRegistry string, keys authn.Keychain) DigestFetcher {
	return DigestFetcher{defaultRegistry, keys}
}
