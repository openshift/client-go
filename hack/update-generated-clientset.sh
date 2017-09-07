#!/bin/bash -e

rm _output/bin/client-gen || true
go build -o _output/bin/client-gen github.com/openshift/client-go/vendor/k8s.io/code-generator/cmd/client-gen

# list of package to generate client set for
packages=(
  github.com/openshift/origin/pkg/template/apis/template
)

function generate_clientset_for() {
  local package="$1";shift
  local group="$1";shift
  local name="$1";shift
  echo "-- Generating ${name} client set for ${package} ..."
  _output/bin/client-gen --clientset-path="github.com/openshift/client-go/${group}" \
             --input-base="${package}"                            \
             --output-base="../../../"                           \
             --clientset-name="${name}"                               \
             --go-header-file=hack/boilerplate.txt                    \
             "$@"
}

verify="${VERIFY:-}"

# remove the old client sets if we're not verifying
if [[ -z "${verify}" ]]; then
  for pkg in "${packages[@]}"; do
    shortGroup=$(basename "${pkg}")
    rm -rf ${shortGroup}
  done
fi

for pkg in "${packages[@]}"; do
  shortGroup=$(basename "${pkg}")
  containingPackage=$(dirname "${pkg}")
  generate_clientset_for "${containingPackage}" "${shortGroup}" "internalclientset" --input=${shortGroup} ${verify} "$@"
  generate_clientset_for "${containingPackage}" "${shortGroup}" "clientset" --input=${shortGroup}/v1 ${verify} "$@"
done