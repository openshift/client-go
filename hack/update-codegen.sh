#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../../../k8s.io/code-generator)}

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/apps \
  github.com/openshift/api \
  "apps:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/authorization \
  github.com/openshift/api \
  "authorization:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/build \
  github.com/openshift/api \
  "build:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/image \
  github.com/openshift/api \
  "image:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/network \
  github.com/openshift/api \
  "network:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/oauth \
  github.com/openshift/api \
  "oauth:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/project \
  github.com/openshift/api \
  "project:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/quota \
  github.com/openshift/api \
  "quota:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/route \
  github.com/openshift/api \
  "route:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/security \
  github.com/openshift/api \
  "security:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/template \
  github.com/openshift/api \
  "template:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
  github.com/openshift/client-go/user \
  github.com/openshift/api \
  "user:v1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt

