module github.com/openshift/client-go

go 1.13

require (
	github.com/openshift/api v0.0.0-20200605231317-fb2a6ca106ae
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v0.18.3
	k8s.io/code-generator v0.18.3
)

replace github.com/openshift/api => github.com/deads2k/api v0.0.0-20200622152052-6eccdf1efede
