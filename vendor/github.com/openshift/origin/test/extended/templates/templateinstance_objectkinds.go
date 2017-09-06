package templates

import (
	"fmt"
	"time"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kapi "k8s.io/kubernetes/pkg/api"

	exutil "github.com/openshift/origin/test/extended/util"
)

// ensure that we can instantiate Kubernetes and OpenShift objects, legacy and
// non-legacy, from a range of API groups.
var _ = g.Describe("[Conformance][templates] templateinstance object kinds test", func() {
	defer g.GinkgoRecover()

	var (
		fixture = exutil.FixturePath("testdata", "templates", "templateinstance_objectkinds.yaml")
		cli     = exutil.NewCLI("templates", exutil.KubeConfigPath())
	)

	g.It("should create objects from varying API groups", func() {
		err := cli.Run("create").Args("-f", fixture).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		// wait for templateinstance controller to do its thing
		err = wait.Poll(time.Second, time.Minute, func() (bool, error) {
			templateinstance, err := cli.TemplateClient().Template().TemplateInstances(cli.Namespace()).Get("templateinstance", metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			for _, c := range templateinstance.Status.Conditions {
				if c.Reason == "Failed" && c.Status == kapi.ConditionTrue {
					return false, fmt.Errorf("failed condition: %s", c.Message)
				}
				if c.Reason == "Created" && c.Status == kapi.ConditionTrue {
					return true, nil
				}
			}

			return false, nil
		})
		o.Expect(err).NotTo(o.HaveOccurred())

		// check everything was created as expected
		_, err = cli.KubeClient().CoreV1().Secrets(cli.Namespace()).Get("secret", metav1.GetOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())

		_, err = cli.KubeClient().AppsV1beta1().Deployments(cli.Namespace()).Get("deployment", metav1.GetOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())

		_, err = cli.Client().Routes(cli.Namespace()).Get("route", metav1.GetOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())

		_, err = cli.Client().Routes(cli.Namespace()).Get("newroute", metav1.GetOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())
	})
})
