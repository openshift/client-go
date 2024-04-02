#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../../../k8s.io/code-generator)}

# TODO(soltysh/dinhxuanvu): this should be removed when bumping to k8s 1.30,
# which already includes these changes, including hack/kube_codegen.sh file:
mv "${CODEGEN_PKG}/kube_codegen.sh" "${CODEGEN_PKG}/kube_codegen.sh.orig"
cp "${SCRIPT_ROOT}/hack/kube_codegen.sh" "${CODEGEN_PKG}/kube_codegen.sh"
function cleanup {
  mv "${CODEGEN_PKG}/kube_codegen.sh.orig" "${CODEGEN_PKG}/kube_codegen.sh"
}
trap cleanup EXIT

source "${CODEGEN_PKG}/kube_codegen.sh"

for group in apiserver apps authorization build cloudnetwork config console helm image imageregistry insights machine monitoring network oauth operator operatorcontrolplane project quota route samples security securityinternal servicecertsigner sharedresource template user; do
  echo "# Processing ${group} ..."
  kube::codegen::gen_client \
      --with-watch \
      --with-applyconfig \
      --applyconfig-name "applyconfigurations" \
      --applyconfig-externals "github.com/openshift/api/operator/v1.OperatorSpec:github.com/openshift/client-go/operator/applyconfigurations/operator/v1,github.com/openshift/api/operator/v1.OperatorStatus:github.com/openshift/client-go/operator/applyconfigurations/operator/v1,github.com/openshift/api/operator/v1.OperatorCondition:github.com/openshift/client-go/operator/applyconfigurations/operator/v1,github.com/openshift/api/operator/v1.GenerationStatus:github.com/openshift/client-go/operator/applyconfigurations/operator/v1" \
      --applyconfig-openapi-schema "vendor/github.com/openshift/api/openapi/openapi.json" \
      --one-input-api "${group}" \
      --output-pkg "github.com/openshift/client-go/${group}" \
      --output-dir "${SCRIPT_ROOT}/${group}" \
      --plural-exceptions "DNS:DNSes,DNSList:DNSList,SecurityContextConstraints:SecurityContextConstraints" \
      --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.txt" \
      "${SCRIPT_ROOT}/vendor/github.com/openshift/api"
done

# machineconfiguration is almost identical to the above call, except for the additional
# --applyconfig-externals value 'k8s.io/api/core/v1.ObjectReference:k8s.io/client-go/applyconfigurations/core/v1'
for group in machineconfiguration; do
  echo "# Processing ${group} ..."
  kube::codegen::gen_client \
      --with-watch \
      --with-applyconfig \
      --applyconfig-name "applyconfigurations" \
      --applyconfig-externals "k8s.io/api/core/v1.ObjectReference:k8s.io/client-go/applyconfigurations/core/v1,github.com/openshift/api/operator/v1.OperatorSpec:github.com/openshift/client-go/operator/applyconfigurations/operator/v1,github.com/openshift/api/operator/v1.OperatorStatus:github.com/openshift/client-go/operator/applyconfigurations/operator/v1,github.com/openshift/api/operator/v1.OperatorCondition:github.com/openshift/client-go/operator/applyconfigurations/operator/v1,github.com/openshift/api/operator/v1.GenerationStatus:github.com/openshift/client-go/operator/applyconfigurations/operator/v1" \
      --applyconfig-openapi-schema "vendor/github.com/openshift/api/openapi/openapi.json" \
      --one-input-api "${group}" \
      --output-pkg "github.com/openshift/client-go/${group}" \
      --output-dir "${SCRIPT_ROOT}/${group}" \
      --plural-exceptions "DNS:DNSes,DNSList:DNSList,SecurityContextConstraints:SecurityContextConstraints" \
      --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.txt" \
      "${SCRIPT_ROOT}/vendor/github.com/openshift/api"
done
