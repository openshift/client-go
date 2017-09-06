package idling

import (
	"time"

	"github.com/openshift/origin/pkg/util/errors"
	exutil "github.com/openshift/origin/test/extended/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func waitForEndpointsAvailable(oc *exutil.CLI, serviceName string) error {
	return wait.Poll(200*time.Millisecond, 3*time.Minute, func() (bool, error) {
		ep, err := oc.KubeClient().CoreV1().Endpoints(oc.Namespace()).Get(serviceName, metav1.GetOptions{})
		// Tolerate NotFound b/c it could take a moment for the endpoints to be created
		if errors.TolerateNotFoundError(err) != nil {
			return false, err
		}

		return (len(ep.Subsets) > 0) && (len(ep.Subsets[0].Addresses) > 0), nil
	})
}

func waitForNoPodsAvailable(oc *exutil.CLI) error {
	return wait.Poll(200*time.Millisecond, 3*time.Minute, func() (bool, error) {
		//ep, err := oc.KubeClient().CoreV1().Endpoints(oc.Namespace()).Get(serviceName)
		pods, err := oc.KubeClient().CoreV1().Pods(oc.Namespace()).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		return len(pods.Items) == 0, nil
	})
}
