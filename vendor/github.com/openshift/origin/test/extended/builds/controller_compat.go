package builds

import (
	"os"

	g "github.com/onsi/ginkgo"

	"github.com/openshift/origin/test/common/build"
	exutil "github.com/openshift/origin/test/extended/util"
)

var _ = g.Describe("[bldcompat][Slow][Compatibility] build controller", func() {
	defer g.GinkgoRecover()
	var (
		oc = exutil.NewCLI("compat-build-controllers", exutil.KubeConfigPath())
	)

	g.JustBeforeEach(func() {
		os.Setenv("OS_TEST_NAMESPACE", oc.Namespace())
	})

	g.Describe("RunBuildControllerTest", func() {
		g.It("should succeed", func() {
			build.RunBuildControllerTest(g.GinkgoT(), oc.AdminClient(), oc.InternalAdminKubeClient())
		})
	})
	g.Describe("RunBuildControllerPodSyncTest", func() {
		g.It("should succeed", func() {
			build.RunBuildControllerPodSyncTest(g.GinkgoT(), oc.AdminClient(), oc.InternalAdminKubeClient())
		})
	})
	g.Describe("RunImageChangeTriggerTest [SkipPrevControllers]", func() {
		g.It("should succeed", func() {
			build.RunImageChangeTriggerTest(g.GinkgoT(), oc.AdminClient())
		})
	})
	g.Describe("RunBuildDeleteTest", func() {
		g.It("should succeed", func() {
			build.RunBuildDeleteTest(g.GinkgoT(), oc.AdminClient(), oc.InternalAdminKubeClient())
		})
	})
	g.Describe("RunBuildRunningPodDeleteTest", func() {
		g.It("should succeed", func() {
			build.RunBuildRunningPodDeleteTest(g.GinkgoT(), oc.AdminClient(), oc.InternalAdminKubeClient())
		})
	})
	g.Describe("RunBuildConfigChangeControllerTest", func() {
		g.It("should succeed", func() {
			build.RunBuildConfigChangeControllerTest(g.GinkgoT(), oc.AdminClient(), oc.InternalAdminKubeClient())
		})
	})
})
