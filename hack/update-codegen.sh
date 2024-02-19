#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../../../k8s.io/code-generator)}

source "${CODEGEN_PKG}/kube_codegen.sh"

# TODO(soltysh):
# verify script

for group in apiserver apps authorization build cloudnetwork image imageregistry oauth project quota route samples security securityinternal template user; do
  kube::codegen::gen_client \
      --with-watch \
      --with-applyconfig \
      --applyconfig-name "applyconfigurations" \
      --input-pkg-root "github.com/openshift/api" \
      --one-input-api "${group}" \
      --output-pkg-root "github.com/openshift/client-go/${group}" \
      --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../.." \
      --plural-exceptions "SecurityContextConstraints:SecurityContextConstraints" \
      --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.txt"
done

for group in machine; do
  kube::codegen::gen_client \
      --with-watch \
      --with-applyconfig \
      --applyconfig-name "applyconfigurations" \
      --input-pkg-root "github.com/openshift/api" \
      --one-input-api "${group}" \
      --output-pkg-root "github.com/openshift/client-go/${group}" \
      --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../.." \
      --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.txt"
done

for group in config console operator monitoring network machineconfiguration; do
  kube::codegen::gen_client \
      --with-watch \
      --with-applyconfig \
      --applyconfig-name "applyconfigurations" \
      --input-pkg-root "github.com/openshift/api" \
      --one-input-api "${group}" \
      --output-pkg-root "github.com/openshift/client-go/${group}" \
      --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../.." \
      --plural-exceptions "DNS:DNSes,DNSList:DNSList" \
      --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.txt"
done

for group in helm; do
  kube::codegen::gen_client \
      --with-watch \
      --with-applyconfig \
      --applyconfig-name "applyconfigurations" \
      --input-pkg-root "github.com/openshift/api" \
      --one-input-api "${group}" \
      --output-pkg-root "github.com/openshift/client-go/${group}" \
      --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../.." \
      --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.txt"
done

for group in servicecertsigner operatorcontrolplane sharedresource insights; do
  kube::codegen::gen_client \
      --with-watch \
      --with-applyconfig \
      --applyconfig-name "applyconfigurations" \
      --input-pkg-root "github.com/openshift/api" \
      --one-input-api "${group}" \
      --output-pkg-root "github.com/openshift/client-go/${group}" \
      --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../.." \
      --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.txt"
done
