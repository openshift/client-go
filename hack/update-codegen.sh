#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=$(dirname ${BASH_SOURCE})/code-generator

verify="${VERIFY:-}"

for group in apps authorization build config console image network oauth project quota route security template user; do
  ${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    ${verify}
done

for group in servicecertsigner; do
  ${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    ${verify}
done

for group in operator; do
  ${CODEGEN_PKG}/generate-groups.sh "client,lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    ${verify}
done
