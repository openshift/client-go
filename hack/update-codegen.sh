#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../../../k8s.io/code-generator)}

verify="${VERIFY:-}"

set -x
# Because go mod sux, we have to fake the vendor for generator in order to be able to build it...
mv ${CODEGEN_PKG}/generate-groups.sh ${CODEGEN_PKG}/generate-groups.sh.orig
sed 's/go install/#GO111MODULE=on go install/g' ${CODEGEN_PKG}/generate-groups.sh.orig > ${CODEGEN_PKG}/generate-groups.sh
# Originally this script doesn't have permissions to run
sed 's/^exec "$(dirname "${BASH_SOURCE\[0\]}")\/generate-internal-groups.sh"/bash "$(dirname "${BASH_SOURCE\[0\]}")\/generate-internal-groups.sh"/g' ${CODEGEN_PKG}/generate-groups.sh.orig > ${CODEGEN_PKG}/generate-groups.sh
# For verification we need to ensure we don't remove files
# TODO (soltysh): this should be properly resolved upstream so that we can get
# rid of the below if condition for verify scripts
if [ ! -z "$verify" ]; then
  mv ${CODEGEN_PKG}/generate-internal-groups.sh ${CODEGEN_PKG}/generate-internal-groups.sh.orig
  sed 's/xargs \-0 rm \-f/xargs -0 echo ""/g' ${CODEGEN_PKG}/generate-internal-groups.sh.orig > ${CODEGEN_PKG}/generate-internal-groups.sh
fi
function cleanup {
  mv ${CODEGEN_PKG}/generate-groups.sh.orig ${CODEGEN_PKG}/generate-groups.sh
  if [ ! -z "$verify" ]; then
    mv ${CODEGEN_PKG}/generate-internal-groups.sh.orig ${CODEGEN_PKG}/generate-internal-groups.sh
  fi
}
trap cleanup EXIT

go install ./${CODEGEN_PKG}/cmd/{defaulter-gen,client-gen,lister-gen,informer-gen,deepcopy-gen}

goos=$(go env GOOS)
goarch=$(go env GOARCH)
APPLYCONFIGURATION_GEN=./_output/bin/${goos}/${goarch}/applyconfiguration-gen
go build -o ${APPLYCONFIGURATION_GEN} ./vendor/k8s.io/code-generator/cmd/applyconfiguration-gen

function codegen::join() { local IFS="$1"; shift; echo "$*"; }

function generateApplyConfiguration(){
    local OUTPUT_PKG="$1"
    local APIS_PKG="$2"
    local GROUPS_WITH_VERSIONS="$3"
    shift 3

    local FQ_APIS=() # e.g. k8s.io/api/apps/v1
    for GVs in ${GROUPS_WITH_VERSIONS}; do
      IFS=: read -r G Vs <<<"${GVs}"

      # enumerate versions
      for V in ${Vs//,/ }; do
        FQ_APIS+=("${APIS_PKG}/${G}/${V}")
      done
    done

   if [ "$OUTPUT_PKG" == "github.com/openshift/client-go/machineconfiguration" ] 
   then
    # TODO(jkyros): this is a temporary hack to ensure proper generation until the MCO can sort
    # out their embedded corev1.ObjectReference in MachineConfigPoolStatusConfiguration, which needs
    # to be resolved by the end of release-4.15 so this hack can be removed
    echo "Generating applyconfigurations specifically for MCO"
    applyconfigurationgen_external_apis_csv="$(codegen::join , "${FQ_APIS[@]}")"
    applyconfigurations_package="${OUTPUT_PKG}/${CLIENTSET_PKG_NAME:-applyconfigurations}"
    ${APPLYCONFIGURATION_GEN}  \
      --output-package "${applyconfigurations_package}" \
      --input-dirs "${applyconfigurationgen_external_apis_csv}" \
      --external-applyconfigurations k8s.io/api/core/v1.ObjectReference:k8s.io/client-go/applyconfigurations/core/v1 \
      --external-applyconfigurations github.com/openshift/api/operator/v1.OperatorSpec:github.com/openshift/client-go/operator/applyconfigurations/operator/v1 \
      --external-applyconfigurations github.com/openshift/api/operator/v1.OperatorStatus:github.com/openshift/client-go/operator/applyconfigurations/operator/v1 \
      --external-applyconfigurations github.com/openshift/api/operator/v1.OperatorCondition:github.com/openshift/client-go/operator/applyconfigurations/operator/v1 \
      --external-applyconfigurations github.com/openshift/api/operator/v1.GenerationStatus:github.com/openshift/client-go/operator/applyconfigurations/operator/v1 \
      "$@"
   else
    echo "Generating applyconfigurations for $OUTPUT_PKG"
    applyconfigurationgen_external_apis_csv="$(codegen::join , "${FQ_APIS[@]}")"
    applyconfigurations_package="${OUTPUT_PKG}/${CLIENTSET_PKG_NAME:-applyconfigurations}"
    ${APPLYCONFIGURATION_GEN}  \
      --output-package "${applyconfigurations_package}" \
      --input-dirs "${applyconfigurationgen_external_apis_csv}" \
      --external-applyconfigurations github.com/openshift/api/operator/v1.OperatorSpec:github.com/openshift/client-go/operator/applyconfigurations/operator/v1 \
      --external-applyconfigurations github.com/openshift/api/operator/v1.OperatorStatus:github.com/openshift/client-go/operator/applyconfigurations/operator/v1 \
      --external-applyconfigurations github.com/openshift/api/operator/v1.OperatorCondition:github.com/openshift/client-go/operator/applyconfigurations/operator/v1 \
      --external-applyconfigurations github.com/openshift/api/operator/v1.GenerationStatus:github.com/openshift/client-go/operator/applyconfigurations/operator/v1 \
      "$@"
   fi


}

# Until we get https://github.com/kubernetes/kubernetes/pull/120877 merged we need to
# explicitly set these two variables which are not defaulted properly in generate-internal-groups.sh
export CLIENTSET_PKG=clientset
export CLIENTSET_NAME=versioned

for group in apiserver apps authorization build cloudnetwork image imageregistry machineconfiguration network oauth project quota route samples security securityinternal template user; do
  bash ${CODEGEN_PKG}/generate-groups.sh "lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
  generateApplyConfiguration \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --openapi-schema ./vendor/github.com/openshift/api/openapi/openapi.json \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
  bash ${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --apply-configuration-package github.com/openshift/client-go/${group}/applyconfigurations \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
done

for group in machine; do
  bash ${CODEGEN_PKG}/generate-groups.sh "lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
  generateApplyConfiguration \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --openapi-schema ./vendor/github.com/openshift/api/openapi/openapi.json \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
  bash ${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --apply-configuration-package github.com/openshift/client-go/${group}/applyconfigurations \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
done

for group in console operator config monitoring; do
  bash ${CODEGEN_PKG}/generate-groups.sh "lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
   generateApplyConfiguration \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --openapi-schema ./vendor/github.com/openshift/api/openapi/openapi.json \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
  bash ${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --apply-configuration-package github.com/openshift/client-go/${group}/applyconfigurations \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
done

for group in helm; do
  bash ${CODEGEN_PKG}/generate-groups.sh "lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
  generateApplyConfiguration \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --openapi-schema ./vendor/github.com/openshift/api/openapi/openapi.json \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
  bash ${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --apply-configuration-package github.com/openshift/client-go/${group}/applyconfigurations \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
done

for group in servicecertsigner operatorcontrolplane sharedresource insights; do
  bash ${CODEGEN_PKG}/generate-groups.sh "lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
  generateApplyConfiguration \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --openapi-schema ./vendor/github.com/openshift/api/openapi/openapi.json \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
  bash ${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --apply-configuration-package github.com/openshift/client-go/${group}/applyconfigurations \
    --trim-path-prefix github.com/openshift/client-go \
    ${verify}
done
