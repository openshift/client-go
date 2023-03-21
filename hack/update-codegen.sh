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
function cleanup {
  mv ${CODEGEN_PKG}/generate-groups.sh.orig ${CODEGEN_PKG}/generate-groups.sh
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

    echo "Generating applyconfigurations"
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
}

for group in apiserver apps authorization build cloudnetwork image imageregistry network oauth project quota route samples security securityinternal template user; do
  bash ${CODEGEN_PKG}/generate-groups.sh "lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    ${verify}
  generateApplyConfiguration \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --openapi-schema ./vendor/github.com/openshift/api/openapi/openapi.json \
    ${verify}
  bash ${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --apply-configuration-package github.com/openshift/client-go/${group}/applyconfigurations \
    ${verify}
done

for group in machine; do
  bash ${CODEGEN_PKG}/generate-groups.sh "lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    ${verify}
  generateApplyConfiguration \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --openapi-schema ./vendor/github.com/openshift/api/openapi/openapi.json \
    ${verify}
  bash ${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --apply-configuration-package github.com/openshift/client-go/${group}/applyconfigurations \
    ${verify}
done

for group in console operator config; do
  bash ${CODEGEN_PKG}/generate-groups.sh "lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    ${verify}
   generateApplyConfiguration \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --openapi-schema ./vendor/github.com/openshift/api/openapi/openapi.json \
    ${verify}
  bash ${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1,v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --apply-configuration-package github.com/openshift/client-go/${group}/applyconfigurations \
    ${verify}
done

for group in helm; do
  bash ${CODEGEN_PKG}/generate-groups.sh "lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    ${verify}
  generateApplyConfiguration \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --openapi-schema ./vendor/github.com/openshift/api/openapi/openapi.json \
    ${verify}
  bash ${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1beta1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --apply-configuration-package github.com/openshift/client-go/${group}/applyconfigurations \
    ${verify}
done

for group in servicecertsigner operatorcontrolplane sharedresource monitoring insights; do
  bash ${CODEGEN_PKG}/generate-groups.sh "lister,informer" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    ${verify}
  generateApplyConfiguration \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --openapi-schema ./vendor/github.com/openshift/api/openapi/openapi.json \
    ${verify}
  bash ${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/openshift/client-go/${group} \
    github.com/openshift/api \
    "${group}:v1alpha1" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.txt \
    --plural-exceptions=DNS:DNSes,DNSList:DNSList,Endpoints:Endpoints,Features:Features,FeaturesList:FeaturesList,SecurityContextConstraints:SecurityContextConstraints \
    --apply-configuration-package github.com/openshift/client-go/${group}/applyconfigurations \
    ${verify}
done
