#!/bin/bash

# this is effectively the e2e from v1.0.0.  It is expected to run forever

# This script tests the high level end-to-end functionality demonstrated
# as part of the examples/sample-app

if [[ -z "$(which iptables)" ]]; then
	echo "IPTables not found - the end-to-end test requires a system with iptables for Kubernetes services."
	exit 1
fi
iptables --list > /dev/null 2>&1
if [ $? -ne 0 ]; then
	sudo iptables --list > /dev/null 2>&1
	if [ $? -ne 0 ]; then
		echo "You do not have iptables or sudo privileges.	Kubernetes services will not work without iptables access.	See https://github.com/kubernetes/kubernetes/issues/1859.	Try 'sudo hack/test-end-to-end.sh'."
		exit 1
	fi
fi

set -o errexit
set -o nounset
set -o pipefail

CONFIG_ROOT_DIR=$(dirname "${BASH_SOURCE}")/config
OS_ROOT=$(dirname "${BASH_SOURCE}")/../../..
source "${OS_ROOT}/test/old-start-configs/v1.0.0/util.sh"

if [[ -z "${BASETMPDIR-}" ]]; then
	TMPDIR="${TMPDIR:-"/tmp"}"
	BASETMPDIR="${TMPDIR}/openshift-e2e-v1.0.0"
	sudo rm -rf "${BASETMPDIR}"
	mkdir -p "${BASETMPDIR}"
fi
SERVER_CONFIG_DIR="${BASETMPDIR}/openshift.local.config"


cp -R $CONFIG_ROOT_DIR/openshift.local.config ${SERVER_CONFIG_DIR}
find ${SERVER_CONFIG_DIR} -name "*.yaml" | xargs sed -i "s|/var/lib/openshift|${BASETMPDIR}|g"


echo "[INFO] Starting end-to-end test"

# Use either the latest release built images, or latest.
if [[ -z "${USE_IMAGES-}" ]]; then
	USE_IMAGES='openshift/origin-${component}:latest'
	if [[ -e "${OS_ROOT}/_output/local/releases/.commit" ]]; then
		COMMIT="$(cat "${OS_ROOT}/_output/local/releases/.commit")"
		USE_IMAGES="openshift/origin-\${component}:${COMMIT}"
	fi
fi

ROUTER_TESTS_ENABLED="${ROUTER_TESTS_ENABLED:-true}"
TEST_ASSETS="${TEST_ASSETS:-false}"

if [[ -z "${BASETMPDIR-}" ]]; then
	TMPDIR="${TMPDIR:-"/tmp"}"
	BASETMPDIR="${TMPDIR}/openshift-e2e"
	sudo rm -rf "${BASETMPDIR}"
	mkdir -p "${BASETMPDIR}"
fi
ETCD_DATA_DIR="${BASETMPDIR}/etcd"
VOLUME_DIR="${BASETMPDIR}/volumes"
FAKE_HOME_DIR="${BASETMPDIR}/openshift.local.home"
LOG_DIR="${LOG_DIR:-${BASETMPDIR}/logs}"
ARTIFACT_DIR="${ARTIFACT_DIR:-${BASETMPDIR}/artifacts}"
mkdir -p $LOG_DIR
mkdir -p $ARTIFACT_DIR

DEFAULT_SERVER_IP=localhost
API_HOST="${API_HOST:-${DEFAULT_SERVER_IP}}"
API_PORT="${API_PORT:-8443}"
API_SCHEME="${API_SCHEME:-https}"
MASTER_ADDR="${API_SCHEME}://${API_HOST}:${API_PORT}"
PUBLIC_MASTER_HOST="${PUBLIC_MASTER_HOST:-${API_HOST}}"
KUBELET_SCHEME="${KUBELET_SCHEME:-https}"
KUBELET_HOST="${KUBELET_HOST:-127.0.0.1}"
KUBELET_PORT="${KUBELET_PORT:-10250}"

SERVER_CONFIG_DIR="${BASETMPDIR}/openshift.local.config"
MASTER_CONFIG_DIR="${SERVER_CONFIG_DIR}/master"
NODE_CONFIG_DIR="${SERVER_CONFIG_DIR}/node-${KUBELET_HOST}"

# use the docker bridge ip address until there is a good way to get the auto-selected address from master
# this address is considered stable
# used as a resolve IP to test routing
CONTAINER_ACCESSIBLE_API_HOST="${CONTAINER_ACCESSIBLE_API_HOST:-172.17.42.1}"

STI_CONFIG_FILE="${LOG_DIR}/stiAppConfig.json"
DOCKER_CONFIG_FILE="${LOG_DIR}/dockerAppConfig.json"
CUSTOM_CONFIG_FILE="${LOG_DIR}/customAppConfig.json"
GO_OUT="${OS_ROOT}/_output/local/bin/$(go env GOHOSTOS)/$(go env GOHOSTARCH)"

# set path so OpenShift is available
export PATH="${GO_OUT}:${PATH}"


##### COPIED FROM NEW VERSIONS OF OUR SCRIPTS
function cleanup_openshift {
	ADMIN_KUBECONFIG="${KUBECONFIG}"
	ETCD_PORT="${ETCD_PORT:-4001}"

	set +e
	dump_container_logs

	echo "[INFO] Dumping etcd contents to ${ARTIFACT_DIR}/etcd_dump.json"
	set_curl_args 0 1
	curl ${clientcert_args} -L "${API_SCHEME}://${API_HOST}:${ETCD_PORT}/v2/keys/?recursive=true" > "${ARTIFACT_DIR}/etcd_dump.json"
	echo

	if [[ -z "${SKIP_TEARDOWN-}" ]]; then
		echo "[INFO] Tearing down test"
		kill_all_processes

		echo "[INFO] Stopping k8s docker containers"; docker ps | awk 'index($NF,"k8s_")==1 { print $1 }' | xargs -l -r docker stop
		if [[ -z "${SKIP_IMAGE_CLEANUP-}" ]]; then
			echo "[INFO] Removing k8s docker containers"; docker ps -a | awk 'index($NF,"k8s_")==1 { print $1 }' | xargs -l -r docker rm
		fi
		set -u
	fi

	delete_large_and_empty_logs

	echo "[INFO] Cleanup complete"
	set -e
}

# dump_container_logs writes container logs to $LOG_DIR
function dump_container_logs()
{
	mkdir -p ${LOG_DIR}

	echo "[INFO] Dumping container logs to ${LOG_DIR}"
	for container in $(docker ps -aq); do
		container_name=$(docker inspect -f "{{.Name}}" $container)
		# strip off leading /
		container_name=${container_name:1}
		if [[ "$container_name" =~ ^k8s_ ]]; then
			pod_name=$(echo $container_name | awk 'BEGIN { FS="[_.]+" }; { print $4 }')
			container_name=${pod_name}-$(echo $container_name | awk 'BEGIN { FS="[_.]+" }; { print $2 }')
		fi
		docker logs "$container" >&"${LOG_DIR}/container-${container_name}.log"
	done
}

# kill_all_processes function will kill all
# all processes created by the test script.
function kill_all_processes()
{
	sudo=
	if type sudo &> /dev/null; then
	sudo=sudo
	fi

	pids=($(jobs -pr))
	for i in ${pids[@]}; do
	ps --ppid=${i} | xargs $sudo kill &> /dev/null
	$sudo kill ${i} &> /dev/null &> /dev/null
	done
}

# delete_large_and_empty_logs deletes empty logs and logs over 20MB
function delete_large_and_empty_logs()
{
	# clean up zero byte log files
	# Clean up large log files so they don't end up on jenkins
	find ${ARTIFACT_DIR} -name *.log -size +20M -exec echo Deleting {} because it is too big. \; -exec rm -f {} \;
	find ${LOG_DIR} -name *.log -size +20M -exec echo Deleting {} because it is too big. \; -exec rm -f {} \;
	find ${LOG_DIR} -name *.log -size 0 -exec echo Deleting {} because it is empty. \; -exec rm -f {} \;
}
##### END CLEANUP COPY

function cleanup()
{
	out=$?
	echo
	if [ $out -ne 0 ]; then
		echo "[FAIL] !!!!! Test Failed !!!!"
	else
		echo "[INFO] Test Succeeded"
	fi
	echo

	cleanup_openshift

	echo "[INFO] Exiting"
	exit $out
}

trap "exit" INT TERM
trap "cleanup" EXIT

function wait_for_app() {
	echo "[INFO] Waiting for app in namespace $1"
	echo "[INFO] Waiting for database pod to start"
	wait_for_command "oc get -n $1 pods -l name=database | grep -i Running" $((60*TIME_SEC))

	echo "[INFO] Waiting for database service to start"
	wait_for_command "oc get -n $1 services | grep database" $((20*TIME_SEC))
	DB_IP=$(oc get -n $1 --output-version=v1 --template="{{ .spec.clusterIP }}" service database)

	echo "[INFO] Waiting for frontend pod to start"
	wait_for_command "oc get -n $1 pods | grep frontend | grep -i Running" $((120*TIME_SEC))

	echo "[INFO] Waiting for frontend service to start"
	wait_for_command "oc get -n $1 services | grep frontend" $((20*TIME_SEC))
	FRONTEND_IP=$(oc get -n $1 --output-version=v1 --template="{{ .spec.clusterIP }}" service frontend)

	echo "[INFO] Waiting for database to start..."
	wait_for_url_timed "http://${DB_IP}:5434" "[INFO] Database says: " $((3*TIME_MIN))

	echo "[INFO] Waiting for app to start..."
	wait_for_url_timed "http://${FRONTEND_IP}:5432" "[INFO] Frontend says: " $((2*TIME_MIN))

	echo "[INFO] Testing app"
	wait_for_command '[[ "$(curl -s -X POST http://${FRONTEND_IP}:5432/keys/foo -d value=1337)" = "Key created" ]]'
	wait_for_command '[[ "$(curl -s http://${FRONTEND_IP}:5432/keys/foo)" = "1337" ]]'
}

# Wait for builds to complete
# $1 namespace
function wait_for_build() {
	echo "[INFO] Waiting for $1 namespace build to complete"
	wait_for_command "oc get -n $1 builds | grep -i complete" $((10*TIME_MIN)) "oc get -n $1 builds | grep -i -e failed -e error"
	BUILD_ID=`oc get -n $1 builds --output-version=v1 --template="{{with index .items 0}}{{.metadata.name}}{{end}}"`
	echo "[INFO] Build ${BUILD_ID} finished"
  # TODO: fix
  set +e
	oc build-logs -n $1 $BUILD_ID > $LOG_DIR/$1build.log
  set -e
}

# Setup
stop_openshift_server
echo "[INFO] `openshift version`"
echo "[INFO] Server logs will be at:    ${LOG_DIR}/openshift.log"
echo "[INFO] Test artifacts will be in: ${ARTIFACT_DIR}"
echo "[INFO] Volumes dir is:            ${VOLUME_DIR}"
echo "[INFO] Config dir is:             ${SERVER_CONFIG_DIR}"
echo "[INFO] Using images:              ${USE_IMAGES}"

# Start All-in-one server and wait for health
echo "[INFO] Create certificates for the OpenShift server"


echo "[INFO] Starting OpenShift server"
sudo env "PATH=${PATH}" OPENSHIFT_PROFILE=web OPENSHIFT_ON_PANIC=crash openshift start \
	--master-config=${MASTER_CONFIG_DIR}/master-config.yaml \
	--node-config=${NODE_CONFIG_DIR}/node-config.yaml \
    --loglevel=4 \
    &> "${LOG_DIR}/openshift.log" &
OS_PID=$!

export HOME="${FAKE_HOME_DIR}"
# This directory must exist so Docker can store credentials in $HOME/.dockercfg
mkdir -p ${FAKE_HOME_DIR}

export KUBECONFIG="${MASTER_CONFIG_DIR}/admin.kubeconfig"
CLUSTER_ADMIN_CONTEXT=$(oc config view --flatten -o template --template='{{index . "current-context"}}')

if [[ "${API_SCHEME}" == "https" ]]; then
	export CURL_CA_BUNDLE="${MASTER_CONFIG_DIR}/ca.crt"
	export CURL_CERT="${MASTER_CONFIG_DIR}/admin.crt"
	export CURL_KEY="${MASTER_CONFIG_DIR}/admin.key"

	# Make oc use ${MASTER_CONFIG_DIR}/admin.kubeconfig, and ignore anything in the running user's $HOME dir
	sudo chmod -R a+rwX "${KUBECONFIG}"
	echo "[INFO] To debug: export KUBECONFIG=$KUBECONFIG"
fi


wait_for_url "${KUBELET_SCHEME}://${KUBELET_HOST}:${KUBELET_PORT}/healthz" "[INFO] kubelet: " 0.5 60
wait_for_url "${API_SCHEME}://${API_HOST}:${API_PORT}/healthz" "apiserver: " 0.25 80
wait_for_url "${API_SCHEME}://${API_HOST}:${API_PORT}/healthz/ready" "apiserver(ready): " 0.25 80
wait_for_url "${API_SCHEME}://${API_HOST}:${API_PORT}/api/v1/nodes/${KUBELET_HOST}" "apiserver(nodes): " 0.25 80

# COMPATIBILITY update the cluster roles and role bindings so that new images can be used.
oc adm policy reconcile-cluster-roles --confirm
oc adm policy reconcile-cluster-role-bindings --confirm
# COMPATIBILITY create a service account for the router
echo '{"kind":"ServiceAccount","apiVersion":"v1","metadata":{"name":"router"}}' | oc create -f -
# COMPATIBILITY add the router SA to the privileged SCC so that it can be use to create the router
oc get scc privileged -o json | sed '/\"users\"/a \"system:serviceaccount:default:router\",' | oc replace scc privileged -f -

# add e2e-user as a viewer for the default namespace so we can see infrastructure pieces appear
openshift admin policy add-role-to-user view e2e-user --namespace=default

# create test project so that this shows up in the console
openshift admin new-project test --description="This is an example project to demonstrate OpenShift v3" --admin="e2e-user"
openshift admin new-project docker --description="This is an example project to demonstrate OpenShift v3" --admin="e2e-user"
openshift admin new-project custom --description="This is an example project to demonstrate OpenShift v3" --admin="e2e-user"
openshift admin new-project cache --description="This is an example project to demonstrate OpenShift v3" --admin="e2e-user"

echo "The console should be available at ${API_SCHEME}://${PUBLIC_MASTER_HOST}:${API_PORT}/console."
echo "Log in as 'e2e-user' to see the 'test' project."

# install the router
echo "[INFO] Installing the router"
# COMPATIBILITY remove --credentials parameter
openshift admin router --create --images="${USE_IMAGES}"

# install the registry. The --mount-host option is provided to reuse local storage.
echo "[INFO] Installing the registry"
# COMPATIBILITY remove --credentials parameter
openshift admin registry --create --images="${USE_IMAGES}"

echo "[INFO] Pre-pulling and pushing ruby-22-centos7"
docker pull centos/ruby-22-centos7:latest
echo "[INFO] Pulled ruby-22-centos7"

echo "[INFO] Waiting for Docker registry pod to start"
# TODO: simplify when #4702 is fixed upstream
wait_for_command '[[ "$(oc get endpoints docker-registry --output-version=v1 --template="{{ if .subsets }}{{ len .subsets }}{{ else }}0{{ end }}" || echo "0")" != "0" ]]' $((5*TIME_MIN))

# services can end up on any IP.	Make sure we get the IP we need for the docker registry
DOCKER_REGISTRY=$(oc get --output-version=v1 --template="{{ .spec.clusterIP }}:{{ with index .spec.ports 0 }}{{ .port }}{{ end }}" service docker-registry)

registry="$(dig @${API_HOST} "docker-registry.default.svc.cluster.local." +short A | head -n 1)"
[[ -n "${registry}" && "${registry}:5000" == "${DOCKER_REGISTRY}" ]]

echo "[INFO] Verifying the docker-registry is up at ${DOCKER_REGISTRY}"
wait_for_url_timed "http://${DOCKER_REGISTRY}/healthz" "[INFO] Docker registry says: " $((2*TIME_MIN))

[ "$(dig @${API_HOST} "docker-registry.default.local." A)" ]

# Client setup (log in as e2e-user and set 'test' as the default project)
# This is required to be able to push to the registry!
echo "[INFO] Logging in as a regular user (e2e-user:pass) with project 'test'..."
oc login -u e2e-user -p pass
[ "$(oc whoami | grep 'e2e-user')" ]
oc project cache
token=$(oc config view --flatten -o template --template='{{with index .users 0}}{{.user.token}}{{end}}')
[[ -n ${token} ]]

echo "[INFO] Docker login as e2e-user to ${DOCKER_REGISTRY}"
docker login -u e2e-user -p ${token} -e e2e-user@openshift.com ${DOCKER_REGISTRY}
echo "[INFO] Docker login successful"

echo "[INFO] Tagging and pushing ruby-22-centos7 to ${DOCKER_REGISTRY}/cache/ruby-22-centos7:latest"
docker tag centos/ruby-22-centos7:latest ${DOCKER_REGISTRY}/cache/ruby-22-centos7:latest
docker push ${DOCKER_REGISTRY}/cache/ruby-22-centos7:latest
echo "[INFO] Pushed ruby-22-centos7"

echo "[INFO] Back to 'default' project with 'admin' user..."
oc project ${CLUSTER_ADMIN_CONTEXT}
[ "$(oc whoami | grep 'system:admin')" ]

# The build requires a dockercfg secret in the builder service account in order
# to be able to push to the registry.  Make sure it exists first.
echo "[INFO] Waiting for dockercfg secrets to be generated in project 'test' before building"
wait_for_command "oc get -n test serviceaccount/builder -o yaml | grep dockercfg > /dev/null" $((60*TIME_SEC))

# Process template and create
echo "[INFO] Submitting application template json for processing..."
oc process -n test -f $CONFIG_ROOT_DIR/../examples/sample-app/application-template-stibuild.json > "${STI_CONFIG_FILE}"
oc process -n docker -f $CONFIG_ROOT_DIR/../examples/sample-app/application-template-dockerbuild.json > "${DOCKER_CONFIG_FILE}"
oc process -n custom -f $CONFIG_ROOT_DIR/../examples/sample-app/application-template-custombuild.json > "${CUSTOM_CONFIG_FILE}"

echo "[INFO] Back to 'test' context with 'e2e-user' user"
oc project test

echo "[INFO] Applying STI application config"
oc create -f "${STI_CONFIG_FILE}"

# Wait for build which should have triggered automatically
echo "[INFO] Starting build from ${STI_CONFIG_FILE} and streaming its logs..."
#oc start-build -n test ruby-sample-build --follow
wait_for_build "test"
wait_for_app "test"

#echo "[INFO] Applying Docker application config"
#oc create -n docker -f "${DOCKER_CONFIG_FILE}"
#echo "[INFO] Invoking generic web hook to trigger new docker build using curl"
#curl -k -X POST $API_SCHEME://$API_HOST:$API_PORT/osapi/v1/namespaces/docker/buildconfigs/ruby-sample-build/webhooks/secret101/generic && sleep 3
#wait_for_build "docker"
#wait_for_app "docker"

#echo "[INFO] Applying Custom application config"
#oc create -n custom -f "${CUSTOM_CONFIG_FILE}"
#echo "[INFO] Invoking generic web hook to trigger new custom build using curl"
#curl -k -X POST $API_SCHEME://$API_HOST:$API_PORT/osapi/v1/namespaces/custom/buildconfigs/ruby-sample-build/webhooks/secret101/generic && sleep 3
#wait_for_build "custom"
#wait_for_app "custom"

echo "[INFO] Back to 'default' project with 'admin' user..."
oc project ${CLUSTER_ADMIN_CONTEXT}

# ensure the router is started
# TODO: simplify when #4702 is fixed upstream
wait_for_command '[[ "$(oc get endpoints router --output-version=v1 --template="{{ if .subsets }}{{ len .subsets }}{{ else }}0{{ end }}" || echo "0")" != "0" ]]' $((5*TIME_MIN))

echo "[INFO] Validating routed app response..."
validate_response "-s -k --resolve www.example.com:443:${CONTAINER_ACCESSIBLE_API_HOST} https://www.example.com" "Hello from OpenShift" 0.2 50

# Remote command execution
echo "[INFO] Validating exec"
registry_pod=$(oc get pod -l deploymentconfig=docker-registry --template='{{(index .items 0).metadata.name}}')
# when running as a restricted pod the registry will run with a pre-allocated
# user in the neighborhood of 1000000+.  Look for a substring of the pre-allocated uid range
oc exec -p ${registry_pod} id | grep 10

# Port forwarding
echo "[INFO] Validating port-forward"
oc port-forward -p ${registry_pod} 5001:5000  &> "${LOG_DIR}/port-forward.log" &
wait_for_url_timed "http://localhost:5001/healthz" "[INFO] Docker registry says: " $((10*TIME_SEC))

# UI e2e tests can be found in assets/test/e2e
if [[ "$TEST_ASSETS" == "true" ]]; then
	echo "[INFO] Running UI e2e tests..."
	pushd ${OS_ROOT}/assets > /dev/null
		grunt test-e2e
	popd > /dev/null
fi
