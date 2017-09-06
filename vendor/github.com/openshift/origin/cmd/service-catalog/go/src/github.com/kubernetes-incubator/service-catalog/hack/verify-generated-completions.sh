#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

echo "===== Verifying Generated Completions ====="

platform="$(os::build::host_platform)"
if [[ "${platform}" != "linux/amd64" ]]; then
  os::log::warning "Completions cannot be verified on non-Linux systems (${platform})"
  exit 0
fi

COMPLETION_ROOT_REL="contrib/completions"
COMPLETION_ROOT="${OS_ROOT}/${COMPLETION_ROOT_REL}"
TMP_COMPLETION_ROOT_REL="_output/verify-generated-completions/"
TMP_COMPLETION_ROOT="${OS_ROOT}/${TMP_COMPLETION_ROOT_REL}/${COMPLETION_ROOT_REL}"

echo "Generating fresh completions..."
if ! output=`${OS_ROOT}/hack/update-generated-completions.sh ${TMP_COMPLETION_ROOT_REL} 2>&1`
then
	echo "FAILURE: Generation of fresh spec failed:"
	echo "$output"
	exit 1
fi


echo "Diffing current completions against freshly generated completions..."
ret=0
diff -Naupr "${COMPLETION_ROOT}" "${TMP_COMPLETION_ROOT}" || ret=$?
rm -rf "${TMP_COMPLETION_ROOT}"
if [[ $ret -eq 0 ]]
then
  echo "SUCCESS: Generated completions up to date."
else
  echo "FAILURE: Generated completions out of date. Please run hack/update-generated-completions.sh"
  exit 1
fi
