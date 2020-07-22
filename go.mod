module github.com/openshift/client-go

go 1.13

require (
	github.com/openshift/api v0.0.0-20200715151710-c8ebadbe7a0b
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.19.0-rc.2
	k8s.io/apimachinery v0.19.0-rc.2
	k8s.io/client-go v0.19.0-rc.2
	k8s.io/code-generator v0.19.0-rc.2
)

replace github.com/openshift/api => github.com/marun/api v0.0.0-20200722054824-ad6e9566fac5
