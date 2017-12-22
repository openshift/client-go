#!/bin/bash -e

readonly REQUIRED_GLIDE_VERSION="v0.13"

function verify_glide_version() {
	if ! command -v glide &> /dev/null; then
		echo "[FATAL] Glide was not found in \$PATH. Please install version ${REQUIRED_GLIDE_VERSION} or newer."
		exit 1
	fi

	local glide_version
	glide_version=($(glide --version))
	if [[ "${glide_version[2]}" < "${REQUIRED_GLIDE_VERSION}" ]]; then
		echo "Detected glide version: ${glide_version[*]}."
		echo "Please install Glide version ${REQUIRED_GLIDE_VERSION} or newer."
		exit 1
	fi
}

verify_glide_version

glide update --strip-vendor

