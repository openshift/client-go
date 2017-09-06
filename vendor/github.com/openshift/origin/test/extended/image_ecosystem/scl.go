package image_ecosystem

import (
	"fmt"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/conditions"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	exutil "github.com/openshift/origin/test/extended/util"
)

func getPodNameForTest(image string, t tc) string {
	return fmt.Sprintf("%s-%s-centos7", image, t.Version)
}

// defineTest will create the gingko test.  This ensures the test
// is created with a local copy of all variables the test will need,
// since the test may not run immediately and may run in parallel with other
// tests, so sharing a variable reference is problematic.  (Sharing the oc client
// is ok for these tests).
func defineTest(image string, t tc, oc *exutil.CLI) {
	g.Describe("returning s2i usage when running the image", func() {
		g.It(fmt.Sprintf("%q should print the usage", t.DockerImageReference), func() {
			g.By(fmt.Sprintf("creating a sample pod for %q", t.DockerImageReference))
			pod := exutil.GetPodForContainer(kapiv1.Container{
				Name:  "test",
				Image: t.DockerImageReference,
			})
			_, err := oc.KubeClient().CoreV1().Pods(oc.Namespace()).Create(pod)
			o.Expect(err).NotTo(o.HaveOccurred())

			err = oc.KubeFramework().WaitForPodRunningSlow(pod.Name)
			if err != nil {
				p, e := oc.KubeClient().CoreV1().Pods(oc.Namespace()).Get(pod.Name, metav1.GetOptions{})
				e2e.Logf("error %v waiting for pod %v: ", p, e)
				o.Expect(err).To(o.Equal(conditions.ErrPodCompleted))
			}

			log, err := oc.KubeClient().CoreV1().Pods(oc.Namespace()).GetLogs(pod.Name, &kapiv1.PodLogOptions{}).DoRaw()
			o.Expect(err).NotTo(o.HaveOccurred())

			o.Expect(string(log)).To(o.ContainSubstring("Sample invocation"))
		})
	})
	g.Describe("using the SCL in s2i images", func() {
		g.It(fmt.Sprintf("%q should be SCL enabled", t.DockerImageReference), func() {
			g.By(fmt.Sprintf("creating a sample pod for %q with /bin/bash -c command", t.DockerImageReference))
			pod := exutil.GetPodForContainer(kapiv1.Container{
				Image:   t.DockerImageReference,
				Name:    "test",
				Command: []string{"/bin/bash", "-c", t.Cmd},
			})

			_, err := oc.KubeClient().CoreV1().Pods(oc.Namespace()).Create(pod)
			o.Expect(err).NotTo(o.HaveOccurred())

			err = oc.KubeFramework().WaitForPodRunningSlow(pod.Name)
			if err != nil {
				p, e := oc.KubeClient().CoreV1().Pods(oc.Namespace()).Get(pod.Name, metav1.GetOptions{})
				e2e.Logf("error %v waiting for pod %v: ", p, e)
				o.Expect(err).To(o.Equal(conditions.ErrPodCompleted))
			}

			log, err := oc.KubeClient().CoreV1().Pods(oc.Namespace()).GetLogs(pod.Name, &kapiv1.PodLogOptions{}).DoRaw()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(string(log)).To(o.ContainSubstring(t.Expected))

			g.By(fmt.Sprintf("creating a sample pod for %q", t.DockerImageReference))
			pod = exutil.GetPodForContainer(kapiv1.Container{
				Image:   t.DockerImageReference,
				Name:    "test",
				Command: []string{"/usr/bin/sleep", "infinity"},
			})
			_, err = oc.KubeClient().CoreV1().Pods(oc.Namespace()).Create(pod)
			o.Expect(err).NotTo(o.HaveOccurred())

			err = oc.KubeFramework().WaitForPodRunningSlow(pod.Name)
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("calling the binary using 'oc exec /bin/bash -c'")
			out, err := oc.Run("exec").Args("-p", pod.Name, "--", "/bin/bash", "-c", t.Cmd).Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(out).Should(o.ContainSubstring(t.Expected))

			g.By("calling the binary using 'oc exec /bin/sh -ic'")
			out, err = oc.Run("exec").Args("-p", pod.Name, "--", "/bin/sh", "-ic", t.Cmd).Output()
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(out).Should(o.ContainSubstring(t.Expected))
		})
	})
}

var _ = g.Describe("[image_ecosystem][Slow] openshift images should be SCL enabled", func() {
	defer g.GinkgoRecover()
	var oc = exutil.NewCLI("s2i-usage", exutil.KubeConfigPath())

	g.JustBeforeEach(func() {
		g.By("waiting for builder service account")
		err := exutil.WaitForBuilderAccount(oc.KubeClient().CoreV1().ServiceAccounts(oc.Namespace()))
		o.Expect(err).NotTo(o.HaveOccurred())
	})

	for image, tcs := range GetTestCaseForImages() {
		for _, t := range tcs {
			defineTest(image, t, oc)
		}
	}
})
