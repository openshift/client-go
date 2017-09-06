#!/bin/bash

# This script generates release zips into _output/releases. It requires the openshift/origin-release
# image to be built prior to executing this command via hack/build-base-images.sh.

# NOTE:   only committed code is built.
source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

function cleanup() {
	return_code=$?
	os::util::describe_return_code "${return_code}"
	exit "${return_code}"
}
trap "cleanup" EXIT

export OS_BUILD_ENV_FROM_ARCHIVE=y
export OS_BUILD_ENV_PRESERVE=_output/local

context="${OS_ROOT}/_output/buildenv-context"

# Clean existing output.
rm -rf "${OS_OUTPUT_RELEASEPATH}"
rm -rf "${context}"
mkdir -p "${context}"
mkdir -p "${OS_OUTPUT}"

container="$( os::build::environment::create /bin/sh -c "OS_ONLY_BUILD_PLATFORMS=${OS_ONLY_BUILD_PLATFORMS-} make build-cross" )"
trap "os::build::environment::cleanup ${container}" EXIT

# Perform the build and release in Docker.
(
  OS_GIT_TREE_STATE=clean # set this because we will be pulling from git archive
  os::build::version::get_vars
  echo "++ Building release ${OS_GIT_VERSION}"
)
os::build::environment::withsource "${container}" "${OS_GIT_COMMIT:-HEAD}"
echo "${OS_GIT_COMMIT}" > "${OS_OUTPUT_RELEASEPATH}/.commit"
