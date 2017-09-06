#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

# Cleanup cluster resources created by this test
(
  set +e
  oc adm policy remove-scc-from-user privileged -z router
  oc delete sa/router -n default
  exit 0
) &>/dev/null

defaultimage="openshift/origin-\${component}:latest"
USE_IMAGES=${USE_IMAGES:-$defaultimage}

os::test::junit::declare_suite_start "cmd/router"
# Test running a router
os::cmd::expect_failure_and_text 'oc adm router --dry-run' 'does not exist'
os::cmd::expect_failure_and_text 'oc adm router --dry-run -o yaml' 'service account "router" is not allowed to access the host network on nodes'
os::cmd::expect_failure_and_text 'oc adm router --dry-run -o yaml' 'name: router'
os::cmd::expect_failure_and_text 'oc adm router --dry-run --stats-port=1937 -o yaml' 'containerPort: 1937'
os::cmd::expect_failure_and_text 'oc adm router --dry-run --host-network=false -o yaml' 'service account "router" is not allowed to access host ports on nodes'
os::cmd::expect_failure_and_text 'oc adm router --dry-run --host-network=false -o yaml' 'hostPort: 1936'
os::cmd::expect_success_and_not_text 'oc adm router --dry-run --host-network=false --host-ports=false -o yaml' 'hostPort: 1936'
os::cmd::expect_failure_and_text 'oc adm router --dry-run --host-network=false --stats-port=1937 -o yaml' 'hostPort: 1937'
os::cmd::expect_failure_and_text 'oc adm router --dry-run --service-account=other -o yaml' 'service account "other" is not allowed to access the host network on nodes'
# set ports internally
os::cmd::expect_failure_and_text 'oc adm router --dry-run --host-network=false -o yaml' 'containerPort: 80'
os::cmd::expect_failure_and_text 'oc adm router --dry-run --host-network=false --ports=80:8080 -o yaml' 'port: 8080'
os::cmd::expect_failure_and_text 'oc adm router --dry-run --host-network=false --ports=80,8443:443 -o yaml' 'targetPort: 8443'
os::cmd::expect_failure_and_text 'oc adm router --dry-run --host-network=false -o yaml' 'hostPort'
os::cmd::expect_success_and_not_text 'oc adm router --dry-run --host-network=false --host-ports=false -o yaml' 'hostPort'
# don't use localhost for liveness probe by default
os::cmd::expect_success_and_not_text "oc adm router --dry-run --host-network=false --host-ports=false -o yaml" 'host: localhost'
# client env vars are optional
os::cmd::expect_success_and_not_text 'oc adm router --dry-run --host-network=false --host-ports=false -o yaml' 'OPENSHIFT_MASTER'
os::cmd::expect_success_and_not_text 'oc adm router --dry-run --host-network=false --host-ports=false --secrets-as-env -o yaml' 'OPENSHIFT_MASTER'
# canonical hostname
os::cmd::expect_success_and_text 'oc adm router --dry-run --host-network=false --host-ports=false --router-canonical-hostname=a.b.c.d -o yaml' 'a.b.c.d'
os::cmd::expect_success_and_text 'oc adm router --dry-run --host-network=false --host-ports=false --router-canonical-hostname=1a.b.c.d -o yaml' '1a.b.c.d'
os::cmd::expect_failure_and_text 'oc adm router --dry-run --host-network=false --host-ports=false --router-canonical-hostname=1a._b.c.d -o yaml' 'error: invalid canonical hostname'
os::cmd::expect_failure_and_text 'oc adm router --dry-run --host-network=false --host-ports=false --router-canonical-hostname=1.2.3.4 -o yaml' 'error: canonical hostname must not be an IP address'
# max_conn
os::cmd::expect_success_and_text 'oc adm router --dry-run --host-network=false --host-ports=false --max-connections=14583 -o yaml' '14583'
# ciphers 
os::cmd::expect_success_and_text 'oc adm router --dry-run --host-network=false --host-ports=false --ciphers=modern -o yaml' 'modern'
# strict-sni
os::cmd::expect_success_and_text 'oc adm router --dry-run --host-network=false --host-ports=false --strict-sni -o yaml' 'ROUTER_STRICT_SNI'

# mount tls crt as secret
os::cmd::expect_success_and_not_text 'oc adm router --dry-run --host-network=false --host-ports=false -o yaml' 'value: /etc/pki/tls/private/tls.crt'
os::cmd::expect_failure_and_text "oc adm router --dry-run --host-network=false --host-ports=false --default-cert=${KUBECONFIG} -o yaml" 'the default cert must contain a private key'
os::cmd::expect_success_and_text "oc adm router --dry-run --host-network=false --host-ports=false --default-cert=test/testdata/router/default_pub_keys.pem -o yaml" 'value: /etc/pki/tls/private/tls.crt'
os::cmd::expect_success_and_text "oc adm router --dry-run --host-network=false --host-ports=false --default-cert=test/testdata/router/default_pub_keys.pem -o yaml" 'tls.key:'
os::cmd::expect_success_and_text "oc adm router --dry-run --host-network=false --host-ports=false --default-cert=test/testdata/router/default_pub_keys.pem -o yaml" 'tls.crt: '
os::cmd::expect_success_and_text "oc adm router --dry-run --host-network=false --host-ports=false --default-cert=test/testdata/router/default_pub_keys.pem -o yaml" 'type: kubernetes.io/tls'
# upgrade the router to have access to host networks
os::cmd::expect_success "oc adm policy add-scc-to-user privileged -z router"
# uses localhost for probes
os::cmd::expect_success_and_text "oc adm router --dry-run -o yaml" 'host: localhost'
os::cmd::expect_success_and_text "oc adm router --dry-run --host-network=false -o yaml" 'hostPort'
os::cmd::expect_failure_and_text "oc adm router --ports=80,8443:443" 'container port 8443 and host port 443 must be equal'

os::cmd::expect_success_and_text "oc adm router -o yaml" 'image:.*-haproxy-router:'
os::cmd::expect_success "oc adm router --images='${USE_IMAGES}'"
os::cmd::expect_success_and_text 'oc adm router' 'service exists'
os::cmd::expect_success_and_text 'oc get dc/router -o yaml' 'readinessProbe'

# delete the router and deployment config, leaving the clusterrolebinding and service account
os::cmd::expect_success_and_text "oc delete svc/router" 'service "router" deleted'
os::cmd::expect_success_and_text "oc delete dc/router" 'deploymentconfig "router" deleted'
# create a router and check for success with a warning about the existing clusterrolebinding
os::cmd::expect_success_and_text "oc adm router" 'warning: clusterrolebindings "router-router-role" already exists'

# only when using hostnetwork should we force the probes to use localhost
os::cmd::expect_success_and_not_text "oc adm router -o yaml --host-network=false" 'host: localhost'
os::cmd::expect_success "oc adm router -o yaml | oc delete -f -"
echo "router: ok"

# test ipfailover
os::cmd::expect_failure_and_text 'oc adm ipfailover --dry-run' 'you must specify at least one virtual IP address'
os::cmd::expect_failure_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --dry-run' 'error: ipfailover could not be created'
os::cmd::expect_success 'oc adm policy add-scc-to-user privileged -z ipfailover'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --dry-run' 'Creating IP failover'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --dry-run' 'Success \(dry run\)'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --dry-run -o yaml' 'name: ipfailover'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --dry-run -o name' 'deploymentconfig/ipfailover'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --dry-run -o yaml' '1.2.3.4'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --iptables-chain="MY_CHAIN" --dry-run -o yaml' 'value: MY_CHAIN'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --check-interval=1177 --dry-run -o yaml' 'value: "1177"'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --check-script="ChkScript.sh" --dry-run -o yaml' 'value: ChkScript.sh'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --notify-script="NotScript.sh" --dry-run -o yaml' 'value: NotScript.sh'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --preemption-strategy="nopreempt" --dry-run -o yaml' 'value: nopreempt'
os::cmd::expect_success_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --dry-run -o yaml --vrrp-id-offset=56' 'hostPort: 63056'
os::cmd::expect_failure_and_text 'oc adm ipfailover --virtual-ips="1.2.3.4" --dry-run -o yaml --vrrp-id-offset=255' 'error: The vrrp-id-offset must be in the range 0..254'
os::cmd::expect_success 'oc adm policy remove-scc-from-user privileged -z ipfailover'

# TODO add tests for normal ipfailover creation
# os::cmd::expect_success_and_text 'oc adm ipfailover' 'deploymentconfig "ipfailover" created'
# os::cmd::expect_failure_and_text 'oc adm ipfailover' 'Error from server: deploymentconfig "ipfailover" already exists'
# os::cmd::expect_success_and_text 'oc adm ipfailover -o name --dry-run | xargs oc delete' 'deleted'
echo "ipfailover: ok"

os::test::junit::declare_suite_end
