package builds

import (
	"fmt"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/origin/test/extended/util"
)

var _ = g.Describe("[builds][Slow] Capabilities should be dropped for s2i builders", func() {
	defer g.GinkgoRecover()
	var (
		s2ibuilderFixture      = exutil.FixturePath("testdata", "s2i-dropcaps", "rootable-ruby")
		rootAccessBuildFixture = exutil.FixturePath("testdata", "s2i-dropcaps", "root-access-build.yaml")
		oc                     = exutil.NewCLI("build-s2i-dropcaps", exutil.KubeConfigPath())
	)

	g.JustBeforeEach(func() {
		g.By("waiting for builder service account")
		err := exutil.WaitForBuilderAccount(oc.KubeClient().Core().ServiceAccounts(oc.Namespace()))
		o.Expect(err).NotTo(o.HaveOccurred())
	})

	g.Describe("s2i build with a rootable builder", func() {
		g.It("should not be able to switch to root with an assemble script", func() {

			g.By("calling oc new-build for rootable-builder")
			err := oc.Run("new-build").Args("--binary", "--name=rootable-ruby").Execute()
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("starting the rootable-ruby build")
			br, _ := exutil.StartBuildAndWait(oc, "rootable-ruby", fmt.Sprintf("--from-dir=%s", s2ibuilderFixture))
			br.AssertSuccess()

			g.By("creating a build that tries to gain root access via su")
			err = oc.Run("create").Args("-f", rootAccessBuildFixture).Execute()
			o.Expect(err).NotTo(o.HaveOccurred())

			g.By("start the root-access-build which attempts root access")
			br2, _ := exutil.StartBuildAndWait(oc, "root-access-build")
			br2.AssertFailure()
		})
	})

})
