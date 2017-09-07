all: generate build
.PHONY: all

build:
	go build github.com/openshift/client-go/template/clientset
.PHONY: build

generate:
	hack/update-generated-clientset.sh
.PHONY: generate