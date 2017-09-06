package templates

import (
	"errors"
	"time"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/test/e2e/framework"

	buildapi "github.com/openshift/origin/pkg/build/apis/build"
	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
	templateapi "github.com/openshift/origin/pkg/template/apis/template"
	exutil "github.com/openshift/origin/test/extended/util"
)

// ensure that template instantiation waits for annotated objects
var _ = g.Describe("[Conformance][templates] templateinstance readiness test", func() {
	defer g.GinkgoRecover()

	var (
		cli              = exutil.NewCLI("templates", exutil.KubeConfigPath())
		template         *templateapi.Template
		templateinstance *templateapi.TemplateInstance
		templatefixture  = exutil.FixturePath("..", "..", "examples", "quickstarts", "cakephp-mysql.json")
	)

	waitSettle := func() (bool, error) {
		var err error

		// must read the templateinstance before the build/dc
		templateinstance, err = cli.TemplateClient().Template().TemplateInstances(cli.Namespace()).Get(templateinstance.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		build, err := cli.Client().Builds(cli.Namespace()).Get("cakephp-mysql-example-1", metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				err = nil
			}
			return false, err
		}

		dc, err := cli.Client().DeploymentConfigs(cli.Namespace()).Get("cakephp-mysql-example", metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				err = nil
			}
			return false, err
		}

		// if the instantiation has settled, quit
		switch build.Status.Phase {
		case buildapi.BuildPhaseCancelled, buildapi.BuildPhaseError, buildapi.BuildPhaseFailed:
			return true, nil

		case buildapi.BuildPhaseComplete:
			var progressing, available *deployapi.DeploymentCondition
			for i, condition := range dc.Status.Conditions {
				switch condition.Type {
				case deployapi.DeploymentProgressing:
					progressing = &dc.Status.Conditions[i]

				case deployapi.DeploymentAvailable:
					available = &dc.Status.Conditions[i]
				}
			}

			if (progressing != nil &&
				progressing.Status == kapi.ConditionTrue &&
				progressing.Reason == deployapi.NewRcAvailableReason &&
				available != nil &&
				available.Status == kapi.ConditionTrue) ||
				(progressing != nil &&
					progressing.Status == kapi.ConditionFalse) {
				return true, nil
			}
		}

		// the build or dc have not settled; the templateinstance must also
		// indicate this

		if templateinstance.HasCondition(templateapi.TemplateInstanceReady, kapi.ConditionTrue) {
			return false, errors.New("templateinstance unexpectedly reported ready")
		}
		if templateinstance.HasCondition(templateapi.TemplateInstanceInstantiateFailure, kapi.ConditionTrue) {
			return false, errors.New("templateinstance unexpectedly reported failure")
		}

		return false, nil
	}

	g.BeforeEach(func() {
		framework.SkipIfProviderIs("gce")

		err := cli.Run("create").Args("-f", templatefixture).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		template, err = cli.TemplateClient().Template().Templates(cli.Namespace()).Get("cakephp-mysql-example", metav1.GetOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())
	})

	g.It("should report ready soon after all annotated objects are ready", func() {
		var err error

		templateinstance = &templateapi.TemplateInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name: "templateinstance",
			},
			Spec: templateapi.TemplateInstanceSpec{
				Template: *template,
			},
		}

		g.By("instantiating the templateinstance")
		templateinstance, err = cli.TemplateClient().Template().TemplateInstances(cli.Namespace()).Create(templateinstance)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("waiting for build and dc to settle")
		err = wait.Poll(time.Second, 20*time.Minute, waitSettle)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("waiting for the templateinstance to indicate ready")
		// in principle, this should happen within 20 seconds
		err = wait.Poll(time.Second, 30*time.Second, func() (bool, error) {
			templateinstance, err = cli.TemplateClient().Template().TemplateInstances(cli.Namespace()).Get(templateinstance.Name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			if templateinstance.HasCondition(templateapi.TemplateInstanceInstantiateFailure, kapi.ConditionTrue) {
				return false, errors.New("templateinstance unexpectedly reported failure")
			}

			return templateinstance.HasCondition(templateapi.TemplateInstanceReady, kapi.ConditionTrue), nil
		})
		o.Expect(err).NotTo(o.HaveOccurred())
	})

	g.It("should report failed soon after an annotated objects has failed", func() {
		var err error

		secret, err := cli.KubeClient().Core().Secrets(cli.Namespace()).Create(&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "secret",
			},
			Data: map[string][]byte{
				"SOURCE_REPOSITORY_URL": []byte("https://bad"),
			},
		})
		o.Expect(err).NotTo(o.HaveOccurred())

		templateinstance = &templateapi.TemplateInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name: "templateinstance",
			},
			Spec: templateapi.TemplateInstanceSpec{
				Template: *template,
				Secret: &kapi.LocalObjectReference{
					Name: secret.Name,
				},
			},
		}

		g.By("instantiating the templateinstance")
		templateinstance, err = cli.TemplateClient().Template().TemplateInstances(cli.Namespace()).Create(templateinstance)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("waiting for build and dc to settle")
		err = wait.Poll(time.Second, 20*time.Minute, waitSettle)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("waiting for the templateinstance to indicate failed")
		// in principle, this should happen within 20 seconds
		err = wait.Poll(time.Second, 30*time.Second, func() (bool, error) {
			templateinstance, err = cli.TemplateClient().Template().TemplateInstances(cli.Namespace()).Get(templateinstance.Name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			if templateinstance.HasCondition(templateapi.TemplateInstanceReady, kapi.ConditionTrue) {
				return false, errors.New("templateinstance unexpectedly reported ready")
			}

			return templateinstance.HasCondition(templateapi.TemplateInstanceInstantiateFailure, kapi.ConditionTrue), nil
		})
		o.Expect(err).NotTo(o.HaveOccurred())
	})
})
