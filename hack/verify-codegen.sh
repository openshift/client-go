#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

"${SCRIPT_ROOT}/hack/update-codegen.sh"

ret=0
git diff --exit-code --quiet || ret=$?
if [[ $ret -ne 0 ]]; then
  echo "Codegen is out of date. Please run hack/update-codegen.sh"
  exit 1
fi
echo "Codegen up to date."

