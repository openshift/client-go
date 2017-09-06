#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

# Cleanup cluster resources created by this test
(
  set +e
  oc delete all,templates --all
  exit 0
) &>/dev/null


os::test::junit::declare_suite_start "cmd/describe"
# This test validates non-duplicate errors when describing an existing resource without a defined describer
os::cmd::expect_success 'oc create -f - << __EOF__
{
  "apiVersion": "v1",
  "involvedObject": {
    "apiVersion": "v1",
    "kind": "Pod",
    "name": "test-pod",
    "namespace": "cmd-describer"
  },
  "kind": "Event",
  "message": "test message",
  "metadata": {
    "name": "test-event"
  }
}
__EOF__
'
os::cmd::try_until_success 'eventnum=$(oc get events | wc -l) && [[ $eventnum -gt 0 ]]'
# resources without describers get a default
os::cmd::expect_success_and_text 'oc describe events' 'Namespace:	cmd-describer'

echo "describer: ok"
os::test::junit::declare_suite_end
