package util

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"

	"github.com/openshift/origin/pkg/util/namer"
)

const (
	// These constants are here to create a name that is short enough to survive chopping by generate name
	maxNameLength             = 63
	randomLength              = 5
	maxSecretPrefixNameLength = maxNameLength - randomLength
)

func GetDockercfgSecretNamePrefix(serviceAccount *kapi.ServiceAccount) string {
	return namer.GetName(serviceAccount.Name, "dockercfg-", maxSecretPrefixNameLength)
}

// GetTokenSecretNamePrefix creates the prefix used for the generated SA token secret. This is compatible with kube up until
// long names, at which point we hash the SA name and leave the "token-" intact.  Upstream clips the value and generates a random
// string.
// TODO fix the upstream implementation to be more like this.
func GetTokenSecretNamePrefix(serviceAccount *kapi.ServiceAccount) string {
	return namer.GetName(serviceAccount.Name, "token-", maxSecretPrefixNameLength)
}

func GetDockercfgSecretNamePrefixV1(serviceAccount *v1.ServiceAccount) string {
	return namer.GetName(serviceAccount.Name, "dockercfg-", maxSecretPrefixNameLength)
}

// GetTokenSecretNamePrefix creates the prefix used for the generated SA token secret. This is compatible with kube up until
// long names, at which point we hash the SA name and leave the "token-" intact.  Upstream clips the value and generates a random
// string.
// TODO fix the upstream implementation to be more like this.
func GetTokenSecretNamePrefixV1(serviceAccount *v1.ServiceAccount) string {
	return namer.GetName(serviceAccount.Name, "token-", maxSecretPrefixNameLength)
}
