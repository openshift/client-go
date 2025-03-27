all: build examples
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/build-machinery-go/make/, \
	golang.mk \
	targets/openshift/deps.mk \
)

EXCLUDE_DIRS := _output/ dependencymagnet/ hack/ vendor/ examples/
GO_PACKAGES :=$(addsuffix ...,$(addprefix ./,$(filter-out $(EXCLUDE_DIRS), $(wildcard */))))
GO_BUILD_PACKAGES :=$(GO_PACKAGES)
GO_BUILD_PACKAGES_EXPANDED :=$(GO_BUILD_PACKAGES)
# LDFLAGS are not needed for dummy builds (saving time on calling git commands)
GO_LD_FLAGS:=

RUNTIME ?= podman
RUNTIME_IMAGE_NAME ?= registry.ci.openshift.org/openshift/release:rhel-9-release-golang-1.24-openshift-4.20

examples:
	go build -o examples/build/app ./examples/build/
.PHONY: examples

verify:
	GOPATH= hack/verify-codegen.sh
.PHONY: verify

update: generate
.PHONY: update

generate:
	GOPATH= hack/update-codegen.sh
.PHONY: generate

verify-with-container:
	$(RUNTIME) run -ti --rm -v $(PWD):/go/src/github.com/openshift/client-go:z -w /go/src/github.com/openshift/client-go $(RUNTIME_IMAGE_NAME) make verify

generate-with-container:
	$(RUNTIME) run -ti --rm -v $(PWD):/go/src/github.com/openshift/client-go:z -w /go/src/github.com/openshift/client-go $(RUNTIME_IMAGE_NAME) make update
