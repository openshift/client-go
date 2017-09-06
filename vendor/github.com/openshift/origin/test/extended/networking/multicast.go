package networking

import (
	"fmt"
	"regexp"

	networkapi "github.com/openshift/origin/pkg/network/apis/network"
	testexutil "github.com/openshift/origin/test/extended/util"
	testutil "github.com/openshift/origin/test/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("[networking] multicast", func() {
	InSingleTenantContext(func() {
		oc := testexutil.NewCLI("multicast", testexutil.KubeConfigPath())
		f := oc.KubeFramework()

		It("should block multicast traffic", func() {
			Expect(testMulticast(f, oc)).NotTo(Succeed())
		})
	})

	InMultiTenantContext(func() {
		oc := testexutil.NewCLI("multicast", testexutil.KubeConfigPath())
		f := oc.KubeFramework()

		It("should block multicast traffic in namespaces where it is disabled", func() {
			Expect(testMulticast(f, oc)).NotTo(Succeed())
		})
		It("should allow multicast traffic in namespaces where it is enabled", func() {
			makeNamespaceMulticastEnabled(f.Namespace)
			Expect(testMulticast(f, oc)).To(Succeed())
		})
	})
})

func makeNamespaceMulticastEnabled(ns *kapiv1.Namespace) {
	client, err := testutil.GetClusterAdminClient(testexutil.KubeConfigPath())
	expectNoError(err)
	netns, err := client.NetNamespaces().Get(ns.Name, metav1.GetOptions{})
	expectNoError(err)
	if netns.Annotations == nil {
		netns.Annotations = make(map[string]string, 1)
	}
	netns.Annotations[networkapi.MulticastEnabledAnnotation] = "true"
	_, err = client.NetNamespaces().Update(netns)
	expectNoError(err)
}

// We run 'omping -c 1 -T 60 -q -q ${ip1} ${ip2} ${ip3}' in each pod:
//   -c 1  : exchange 1 multicast packet with each peer and then exit
//   -T 60 : time out and exit after 60 seconds no matter what
//   -q -q : extra quiet, only print final status
//
// (Since we need to pass all three pod IPs to each omping command, we launch the pods
// with the command "sleep 1000" first and then use "oc exec" to run omping.)
//
// Each omping instance will try to send a unicast packet to other instance until it
// succeeds or times out. Once it succeeds, it will then send a multicast packet to that
// peer, and expect to receive a multicast packet. After it has communicated with both
// peers, it will exit.
//
// The 60-second timeout only gets hit if unicast communication fails; if unicast works
// but multicast doesn't then omping will fail within a few seconds of exchanging unicast
// packets.
//
// The output looks like:
//
//   10.130.0.3 :   unicast, xmt/rcv/%loss = 1/1/0%, min/avg/max/std-dev = 0.046/0.046/0.046/0.000
//   10.130.0.3 : multicast, xmt/rcv/%loss = 1/1/0%, min/avg/max/std-dev = 0.068/0.068/0.068/0.000
//   10.129.0.2 :   unicast, xmt/rcv/%loss = 1/1/0%, min/avg/max/std-dev = 0.066/0.066/0.066/0.000
//   10.129.0.2 : multicast, xmt/rcv/%loss = 1/1/0%, min/avg/max/std-dev = 0.095/0.095/0.095/0.000
//
// (or, on failure, "multicast, xmt/rcv/%loss = 1/0/100%, ...")

func testMulticast(f *e2e.Framework, oc *testexutil.CLI) error {
	nodes := e2e.GetReadySchedulableNodesOrDie(f.ClientSet)
	if len(nodes.Items) == 1 {
		e2e.Skipf("Only one node is available in this environment")
	}

	var pod, ip, out [3]string
	var err [3]error
	var ch [3]chan struct{}
	var matchIP [3]*regexp.Regexp

	for i := range pod {
		pod[i] = fmt.Sprintf("multicast-%d", i)
		ip[i], err[i] = launchTestMulticastPod(f, nodes.Items[i/2].Name, pod[i])
		if err[i] != nil {
			return err[i]
		}
		var zero int64
		defer f.ClientSet.CoreV1().Pods(f.Namespace.Name).Delete(pod[i], &metav1.DeleteOptions{GracePeriodSeconds: &zero})
		matchIP[i] = regexp.MustCompile(ip[i] + ".*multicast.*1/1/0%")
		ch[i] = make(chan struct{})
	}

	for i := range pod {
		i := i
		go func() {
			out[i], err[i] = oc.Run("exec").Args(pod[i], "--", "omping", "-c", "1", "-T", "60", "-q", "-q", ip[0], ip[1], ip[2]).Output()
			close(ch[i])
		}()
	}
	for i := range pod {
		<-ch[i]
		if err[i] != nil {
			return err[i]
		}
		for j := range pod {
			if i != j {
				if !matchIP[j].MatchString(out[i]) {
					return fmt.Errorf("pod %d failed to send multicast to pod %d", i, j)
				}
			}
		}
	}

	return nil
}

func launchTestMulticastPod(f *e2e.Framework, nodeName string, podName string) (string, error) {
	contName := fmt.Sprintf("%s-container", podName)
	pod := &kapiv1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: kapiv1.PodSpec{
			Containers: []kapiv1.Container{
				{
					Name:    contName,
					Image:   "openshift/test-multicast",
					Command: []string{"sleep", "1000"},
				},
			},
			NodeName:      nodeName,
			RestartPolicy: kapiv1.RestartPolicyNever,
		},
	}
	podClient := f.ClientSet.CoreV1().Pods(f.Namespace.Name)
	_, err := podClient.Create(pod)
	expectNoError(err)

	podIP := ""
	err = waitForPodCondition(f.ClientSet, f.Namespace.Name, podName, "running", podStartTimeout, func(pod *kapiv1.Pod) (bool, error) {
		podIP = pod.Status.PodIP
		return podIP != "", nil
	})
	return podIP, err
}
