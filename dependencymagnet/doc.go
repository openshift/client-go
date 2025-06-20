//go:build tools
// +build tools

// go mod won't pull in code that isn't depended upon, but we have some code we don't depend on from code that must be included
// for our build to work.
package dependencymagnet

import (
	// Also pulls in every dependency in api/install.go, which includes all
	// our client apis
	_ "github.com/openshift/api"

	// The openapi package is not pulled in by api/install.go
	_ "github.com/openshift/api/openapi"

	_ "github.com/openshift/build-machinery-go"
	_ "github.com/spf13/pflag"
	_ "k8s.io/code-generator"
	_ "k8s.io/code-generator/cmd/applyconfiguration-gen"
)
