module github.com/openshift/client-go

go 1.13

require (
	github.com/openshift/api v0.0.0-20200521101457-60c476765272
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v0.18.3
	k8s.io/code-generator v0.18.3
)

replace github.com/openshift/api => github.com/sanchezl/api v0.0.0-20200518144445-c4f5eb11d75b
