module github.com/openshift/client-go

go 1.16

require (
	github.com/openshift/api v0.0.0-20220413151724-070bf8b26f73
	github.com/spf13/pflag v1.0.5
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/code-generator v0.23.0
)

replace github.com/openshift/api => github.com/damdo/api v0.0.0-20220419104543-f3d1511cf2fa
