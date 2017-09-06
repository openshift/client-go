#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

# This test validates the diagnostics command

# available diagnostics (2016-09-01):
# AnalyzeLogs ClusterRegistry ClusterRoleBindings ClusterRoles ClusterRouter ConfigContexts DiagnosticPod MasterConfigCheck MasterNode NetworkCheck NodeConfigCheck NodeDefinitions UnitStatus MetricsApiProxy ServiceExternalIPs
# Without things feeding into systemd, AnalyzeLogs and UnitStatus are irrelevant.
# The rest should be included in some fashion.

os::test::junit::declare_suite_start "cmd/diagnostics"

# validate config that was generated
os::cmd::expect_success "oc adm diagnostics MasterConfigCheck --master-config='${MASTER_CONFIG_DIR}/master-config.yaml'"
os::cmd::expect_success "oc adm diagnostics NodeConfigCheck --node-config='${NODE_CONFIG_DIR}/node-config.yaml'"
# breaking the config fails the validation check
cp "${MASTER_CONFIG_DIR}/master-config.yaml" "${BASETMPDIR}/master-config-broken.yaml"
os::util::sed '7,12d' "${BASETMPDIR}/master-config-broken.yaml"
os::cmd::expect_failure_and_text "oc adm diagnostics MasterConfigCheck --master-config='${BASETMPDIR}/master-config-broken.yaml'" 'ERROR'

cp "${NODE_CONFIG_DIR}/node-config.yaml" "${BASETMPDIR}/node-config-broken.yaml"
os::util::sed '5,10d' "${BASETMPDIR}/node-config-broken.yaml"
os::cmd::expect_failure_and_text "oc adm diagnostics NodeConfigCheck --node-config='${BASETMPDIR}/node-config-broken.yaml'" 'ERROR'

os::cmd::expect_success 'oc adm policy reconcile-cluster-roles --additive-only=false --confirm'

os::cmd::expect_success 'oc adm diagnostics ClusterRoleBindings ClusterRoles ConfigContexts '
# DiagnosticPod can't run without Docker, would just time out. Exercise flags instead.
os::cmd::expect_success "oc adm diagnostics DiagnosticPod --prevent-modification --images=foo"
os::cmd::expect_success "oc adm diagnostics MasterConfigCheck NodeConfigCheck ServiceExternalIPs --master-config=${MASTER_CONFIG_DIR}/master-config.yaml --node-config=${NODE_CONFIG_DIR}/node-config.yaml"
os::cmd::expect_failure_and_text 'oc adm diagnostics ClusterRegistry' "DClu1006 from diagnostic ClusterRegistry"
# MasterNode fails in test, possibly because the hostname doesn't resolve? Disabled
#os::cmd::expect_success_and_text 'oc adm diagnostics MasterNode'  'Network plugin does not require master to also run node'
# ClusterRouter fails differently depending on whether other tests have run first, so don't test for specific error
# no ordering allowed
#os::cmd::expect_failure 'oc adm diagnostics ClusterRouter' # "DClu2001 from diagnostic ClusterRouter"
os::cmd::expect_failure 'oc adm diagnostics NodeDefinitions'
os::cmd::expect_failure_and_text 'oc adm diagnostics FakeDiagnostic AlsoMissing' 'No requested diagnostics are available: requested=FakeDiagnostic AlsoMissing'
os::cmd::expect_failure_and_text 'oc adm diagnostics AnalyzeLogs AlsoMissing' 'Not all requested diagnostics are available: missing=AlsoMissing requested=AnalyzeLogs AlsoMissing available='
os::cmd::expect_success_and_text 'oc adm diagnostics MetricsApiProxy'  'Skipping diagnostic: MetricsApiProxy'
os::cmd::expect_success_and_text 'oc adm diagnostics NetworkCheck --prevent-modification' 'Skipping diagnostic: NetworkCheck'

# openshift ex diagnostics is deprecated but not removed. Make sure it works until we consciously remove it.
os::cmd::expect_success 'openshift ex diagnostics ClusterRoleBindings ClusterRoles ConfigContexts '
echo "diagnostics: ok"
os::test::junit::declare_suite_end
