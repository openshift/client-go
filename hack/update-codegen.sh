#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

# Get the location of the go packages on disk. These will normally be in the
# vendor directory but may be somewhere else if, e.g. you've overridden it
# locally.
OPENSHIFT_API_PKG=$(go list -f '{{.Dir}}' github.com/openshift/api)
CODEGEN_PKG=$(go list -f '{{.Dir}}' k8s.io/code-generator)

source "${CODEGEN_PKG}/kube_codegen.sh"

for group in apiserver apps authorization build cloudnetwork config console helm image imageregistry insights machine monitoring network oauth operator operatorcontrolplane operatoringress project quota route samples security securityinternal servicecertsigner sharedresource template user; do
  echo "# Processing ${group} ..."
  kube::codegen::gen_client \
      --with-watch \
      --with-applyconfig \
      --applyconfig-name "applyconfigurations" \
      --applyconfig-externals "github.com/openshift/api/operator/v1.OperatorSpec:github.com/openshift/client-go/operator/applyconfigurations/operator/v1,github.com/openshift/api/operator/v1.OperatorStatus:github.com/openshift/client-go/operator/applyconfigurations/operator/v1,github.com/openshift/api/operator/v1.OperatorCondition:github.com/openshift/client-go/operator/applyconfigurations/operator/v1,github.com/openshift/api/operator/v1.GenerationStatus:github.com/openshift/client-go/operator/applyconfigurations/operator/v1" \
      --applyconfig-openapi-schema "${OPENSHIFT_API_PKG}/openapi/openapi.json" \
      --one-input-api "${group}" \
      --output-pkg "github.com/openshift/client-go/${group}" \
      --output-dir "${SCRIPT_ROOT}/${group}" \
      --plural-exceptions "DNS:DNSes,DNSList:DNSList,SecurityContextConstraints:SecurityContextConstraints" \
      --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.txt" \
      "${OPENSHIFT_API_PKG}"
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
      --applyconfig-openapi-schema "${OPENSHIFT_API_PKG}/openapi/openapi.json" \
      --one-input-api "${group}" \
      --output-pkg "github.com/openshift/client-go/${group}" \
      --output-dir "${SCRIPT_ROOT}/${group}" \
      --plural-exceptions "DNS:DNSes,DNSList:DNSList,SecurityContextConstraints:SecurityContextConstraints" \
      --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.txt" \
      "${OPENSHIFT_API_PKG}"
done
