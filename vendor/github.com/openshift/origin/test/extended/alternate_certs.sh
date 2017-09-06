#!/bin/bash
#
# This scripts starts the OpenShift server with custom TLS certs, and verifies generated kubeconfig files can be used to talk to it.
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"

os::cleanup::tmpdir
os::util::environment::setup_all_server_vars

function cleanup() {
	return_code=$?
	os::test::junit::generate_report
	os::cleanup::all
	os::util::describe_return_code "${return_code}"
	exit "${return_code}"
}
trap "cleanup" EXIT

os::log::info "Starting server as distinct processes"
os::log::info "`openshift version`"
os::log::info "Server logs will be at:    ${LOG_DIR}/openshift.log"
os::log::info "Test artifacts will be in: ${ARTIFACT_DIR}"
os::log::info "Config dir is:             ${SERVER_CONFIG_DIR}"

# Allow setting $JUNIT_REPORT to toggle output behavior
if [[ -n "${JUNIT_REPORT:-}" ]]; then
	export JUNIT_REPORT_OUTPUT="${LOG_DIR}/raw_test_output.log"
fi

mkdir -p ${LOG_DIR}

os::log::info "Scan of OpenShift related processes already up via ps -ef	| grep openshift : "
ps -ef | grep openshift

mkdir -p "${SERVER_CONFIG_DIR}"
pushd "${SERVER_CONFIG_DIR}"

os::test::junit::declare_suite_start "extended/alternate_certs"

# Make custom CA and server cert
os::cmd::expect_success 'oc adm ca create-signer-cert --overwrite=true --cert=master/custom-ca.crt --key=master/custom-ca.key --serial=master/custom-ca.txt --name=my-custom-ca@`date +%s`'
os::cmd::expect_success 'oc adm ca create-server-cert --cert=master/custom.crt --key=master/custom.key --hostnames=localhost,customhost.com --signer-cert=master/custom-ca.crt --signer-key=master/custom-ca.key --signer-serial=master/custom-ca.txt'

# Create master/node configs
os::cmd::expect_success "openshift start --master=https://localhost:${API_PORT} --write-config=. --hostname=mynode --etcd-dir=./etcd --certificate-authority=master/custom-ca.crt"

# Don't try this at home.  We don't have flags for setting etcd ports in the config, but we want deconflicted ones.  Use sed to replace defaults in a completely unsafe way
os::util::sed "s/:4001$/:${ETCD_PORT}/g" master/master-config.yaml
os::util::sed "s/:7001$/:${ETCD_PEER_PORT}/g" master/master-config.yaml
# replace top-level namedCertificates config
os::util::sed 's#^  namedCertificates: null#  namedCertificates: [{"certFile":"custom.crt","keyFile":"custom.key","names":["localhost"]}]#' master/master-config.yaml

# Start master
OPENSHIFT_PROFILE=web OPENSHIFT_ON_PANIC=crash openshift start master \
 --config=master/master-config.yaml \
 --loglevel=4 \
&>"${LOG_DIR}/openshift.log" &
OS_PID=$!

# Wait for the server to be up
os::cmd::try_until_success "oc whoami --config=master/admin.kubeconfig"

# Verify the server is serving with the custom and internal CAs, and that the generated ca-bundle.crt works for both
os::cmd::expect_success_and_text "curl -vvv https://localhost:${API_PORT} --cacert master/ca-bundle.crt -s 2>&1" 'my-custom-ca'
os::cmd::expect_success_and_text "curl -vvv https://127.0.0.1:${API_PORT} --cacert master/ca-bundle.crt -s 2>&1" 'openshift-signer'

# Verify kubeconfigs have connectivity to hosts serving with custom and generated certs
os::cmd::expect_success_and_text "oc whoami --config=master/admin.kubeconfig"                                        'system:admin'
os::cmd::expect_success_and_text "oc whoami --config=master/admin.kubeconfig --server=https://localhost:${API_PORT}" 'system:admin'
os::cmd::expect_success_and_text "oc whoami --config=master/admin.kubeconfig --server=https://127.0.0.1:${API_PORT}" 'system:admin'

os::cmd::expect_success_and_text "oc whoami --config=master/openshift-master.kubeconfig"                                        'system:openshift-master'
os::cmd::expect_success_and_text "oc whoami --config=master/openshift-master.kubeconfig --server=https://localhost:${API_PORT}" 'system:openshift-master'
os::cmd::expect_success_and_text "oc whoami --config=master/openshift-master.kubeconfig --server=https://127.0.0.1:${API_PORT}" 'system:openshift-master'

os::cmd::expect_success_and_text "oc whoami --config=node-mynode/node.kubeconfig"                                        'system:node:mynode'
os::cmd::expect_success_and_text "oc whoami --config=node-mynode/node.kubeconfig --server=https://localhost:${API_PORT}" 'system:node:mynode'
os::cmd::expect_success_and_text "oc whoami --config=node-mynode/node.kubeconfig --server=https://127.0.0.1:${API_PORT}" 'system:node:mynode'

os::test::junit::declare_suite_end

kill $OS_PID

popd
