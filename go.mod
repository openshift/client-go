module github.com/openshift/client-go

go 1.16

require (
	github.com/openshift/api v0.0.0-20220823132416-944d64f14f39
	github.com/spf13/pflag v1.0.5
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	k8s.io/api v0.24.0
	k8s.io/apimachinery v0.24.0
	k8s.io/client-go v0.24.0
	k8s.io/code-generator v0.24.0
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3
)

replace k8s.io/code-generator => github.com/openshift/kubernetes-code-generator v0.0.0-20220822200235-042483082c5e

replace github.com/openshift/api => github.com/deads2k/api v0.0.0-20220824122603-eb2ebe1adab1

replace k8s.io/kube-openapi => github.com/deads2k/kube-openapi v0.0.0-20220824152811-e54777fe3dde
