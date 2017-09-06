package integration

import (
	"testing"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	buildapi "github.com/openshift/origin/pkg/build/apis/build"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/server/bootstrappolicy"
	imageapi "github.com/openshift/origin/pkg/image/apis/image"
	"github.com/openshift/origin/pkg/oc/admin/policy"
	testutil "github.com/openshift/origin/test/util"
	testserver "github.com/openshift/origin/test/util/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watchapi "k8s.io/apimachinery/pkg/watch"
	kapi "k8s.io/kubernetes/pkg/api"
)

const (
	streamName = "test-image-trigger-repo"
	tag        = "latest"
)

func TestSimpleImageChangeBuildTriggerFromImageStreamTagSTI(t *testing.T) {
	projectAdminClient, _, fn := setup(t)
	defer fn()
	imageStream := mockImageStream2(tag)
	imageStreamMapping := mockImageStreamMapping(imageStream.Name, "someimage", tag, "registry:8080/openshift/test-image-trigger:"+tag)
	strategy := stiStrategy("ImageStreamTag", streamName+":"+tag)
	config := imageChangeBuildConfig("sti-imagestreamtag", strategy)
	runTest(t, "SimpleImageChangeBuildTriggerFromImageStreamTagSTI", projectAdminClient, imageStream, imageStreamMapping, config, tag)
}

func TestSimpleImageChangeBuildTriggerFromImageStreamTagSTIWithConfigChange(t *testing.T) {
	projectAdminClient, _, fn := setup(t)
	defer fn()
	imageStream := mockImageStream2(tag)
	imageStreamMapping := mockImageStreamMapping(imageStream.Name, "someimage", tag, "registry:8080/openshift/test-image-trigger:"+tag)
	strategy := stiStrategy("ImageStreamTag", streamName+":"+tag)
	config := imageChangeBuildConfigWithConfigChange("sti-imagestreamtag", strategy)
	runTest(t, "SimpleImageChangeBuildTriggerFromImageStreamTagSTI", projectAdminClient, imageStream, imageStreamMapping, config, tag)
}

func TestSimpleImageChangeBuildTriggerFromImageStreamTagDocker(t *testing.T) {
	projectAdminClient, _, fn := setup(t)
	defer fn()
	imageStream := mockImageStream2(tag)
	imageStreamMapping := mockImageStreamMapping(imageStream.Name, "someimage", tag, "registry:8080/openshift/test-image-trigger:"+tag)
	strategy := dockerStrategy("ImageStreamTag", streamName+":"+tag)
	config := imageChangeBuildConfig("docker-imagestreamtag", strategy)
	runTest(t, "SimpleImageChangeBuildTriggerFromImageStreamTagDocker", projectAdminClient, imageStream, imageStreamMapping, config, tag)
}

func TestSimpleImageChangeBuildTriggerFromImageStreamTagDockerWithConfigChange(t *testing.T) {
	projectAdminClient, _, fn := setup(t)
	defer fn()
	imageStream := mockImageStream2(tag)
	imageStreamMapping := mockImageStreamMapping(imageStream.Name, "someimage", tag, "registry:8080/openshift/test-image-trigger:"+tag)
	strategy := dockerStrategy("ImageStreamTag", streamName+":"+tag)
	config := imageChangeBuildConfigWithConfigChange("docker-imagestreamtag", strategy)
	runTest(t, "SimpleImageChangeBuildTriggerFromImageStreamTagDocker", projectAdminClient, imageStream, imageStreamMapping, config, tag)
}

func TestSimpleImageChangeBuildTriggerFromImageStreamTagCustom(t *testing.T) {
	projectAdminClient, clusterAdminClient, fn := setup(t)
	defer fn()

	clusterRoleBindingAccessor := policy.NewClusterRoleBindingAccessor(clusterAdminClient)
	subjects := []kapi.ObjectReference{
		{
			Kind: authorizationapi.SystemGroupKind,
			Name: bootstrappolicy.AuthenticatedGroup,
		},
	}
	options := policy.RoleModificationOptions{
		RoleNamespace:       testutil.Namespace(),
		RoleName:            bootstrappolicy.BuildStrategyCustomRoleName,
		RoleBindingAccessor: clusterRoleBindingAccessor,
		Subjects:            subjects,
	}
	options.AddRole()

	if err := testutil.WaitForPolicyUpdate(projectAdminClient, testutil.Namespace(), "create", buildapi.Resource(authorizationapi.CustomBuildResource), true); err != nil {
		t.Fatal(err)
	}

	imageStream := mockImageStream2(tag)
	imageStreamMapping := mockImageStreamMapping(imageStream.Name, "someimage", tag, "registry:8080/openshift/test-image-trigger:"+tag)
	strategy := customStrategy("ImageStreamTag", streamName+":"+tag)
	config := imageChangeBuildConfig("custom-imagestreamtag", strategy)
	runTest(t, "SimpleImageChangeBuildTriggerFromImageStreamTagCustom", projectAdminClient, imageStream, imageStreamMapping, config, tag)
}

func TestSimpleImageChangeBuildTriggerFromImageStreamTagCustomWithConfigChange(t *testing.T) {
	projectAdminClient, _, fn := setup(t)
	defer fn()

	clusterAdminClient, err := testutil.GetClusterAdminClient(testutil.GetBaseDir() + "/openshift.local.config/master/admin.kubeconfig")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	clusterRoleBindingAccessor := policy.NewClusterRoleBindingAccessor(clusterAdminClient)
	subjects := []kapi.ObjectReference{
		{
			Kind: authorizationapi.SystemGroupKind,
			Name: bootstrappolicy.AuthenticatedGroup,
		},
	}
	options := policy.RoleModificationOptions{
		RoleNamespace:       testutil.Namespace(),
		RoleName:            bootstrappolicy.BuildStrategyCustomRoleName,
		RoleBindingAccessor: clusterRoleBindingAccessor,
		Subjects:            subjects,
	}
	options.AddRole()

	if err := testutil.WaitForPolicyUpdate(projectAdminClient, testutil.Namespace(), "create", buildapi.Resource(authorizationapi.CustomBuildResource), true); err != nil {
		t.Fatal(err)
	}

	imageStream := mockImageStream2(tag)
	imageStreamMapping := mockImageStreamMapping(imageStream.Name, "someimage", tag, "registry:8080/openshift/test-image-trigger:"+tag)
	strategy := customStrategy("ImageStreamTag", streamName+":"+tag)
	config := imageChangeBuildConfigWithConfigChange("custom-imagestreamtag", strategy)
	runTest(t, "SimpleImageChangeBuildTriggerFromImageStreamTagCustom", projectAdminClient, imageStream, imageStreamMapping, config, tag)
}

func dockerStrategy(kind, name string) buildapi.BuildStrategy {
	return buildapi.BuildStrategy{
		DockerStrategy: &buildapi.DockerBuildStrategy{
			From: &kapi.ObjectReference{
				Kind: kind,
				Name: name,
			},
		},
	}
}
func stiStrategy(kind, name string) buildapi.BuildStrategy {
	return buildapi.BuildStrategy{
		SourceStrategy: &buildapi.SourceBuildStrategy{
			From: kapi.ObjectReference{
				Kind: kind,
				Name: name,
			},
		},
	}
}
func customStrategy(kind, name string) buildapi.BuildStrategy {
	return buildapi.BuildStrategy{
		CustomStrategy: &buildapi.CustomBuildStrategy{
			From: kapi.ObjectReference{
				Kind: kind,
				Name: name,
			},
		},
	}
}

func imageChangeBuildConfig(name string, strategy buildapi.BuildStrategy) *buildapi.BuildConfig {
	return &buildapi.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testutil.Namespace(),
			Labels:    map[string]string{"testlabel": "testvalue"},
		},
		Spec: buildapi.BuildConfigSpec{

			RunPolicy: buildapi.BuildRunPolicyParallel,
			CommonSpec: buildapi.CommonSpec{
				Source: buildapi.BuildSource{
					Git: &buildapi.GitBuildSource{
						URI: "git://github.com/openshift/ruby-hello-world.git",
					},
					ContextDir: "contextimage",
				},
				Strategy: strategy,
				Output: buildapi.BuildOutput{
					To: &kapi.ObjectReference{
						Kind: "ImageStreamTag",
						Name: "test-image-trigger-repo:outputtag",
					},
				},
			},
			Triggers: []buildapi.BuildTriggerPolicy{
				{
					Type:        buildapi.ImageChangeBuildTriggerType,
					ImageChange: &buildapi.ImageChangeTrigger{},
				},
			},
		},
	}
}
func imageChangeBuildConfigWithConfigChange(name string, strategy buildapi.BuildStrategy) *buildapi.BuildConfig {
	bc := imageChangeBuildConfig(name, strategy)
	bc.Spec.Triggers = append(bc.Spec.Triggers, buildapi.BuildTriggerPolicy{Type: buildapi.ConfigChangeBuildTriggerType})
	return bc
}

func mockImageStream2(tag string) *imageapi.ImageStream {
	return &imageapi.ImageStream{
		ObjectMeta: metav1.ObjectMeta{Name: "test-image-trigger-repo"},

		Spec: imageapi.ImageStreamSpec{
			DockerImageRepository: "registry:8080/openshift/test-image-trigger",
			Tags: map[string]imageapi.TagReference{
				tag: {
					From: &kapi.ObjectReference{
						Kind: "DockerImage",
						Name: "registry:8080/openshift/test-image-trigger:" + tag,
					},
				},
			},
		},
	}
}

func mockImageStreamMapping(stream, image, tag, reference string) *imageapi.ImageStreamMapping {
	// create a mapping to an image that doesn't exist
	return &imageapi.ImageStreamMapping{
		ObjectMeta: metav1.ObjectMeta{Name: stream},
		Tag:        tag,
		Image: imageapi.Image{
			ObjectMeta: metav1.ObjectMeta{
				Name: image,
			},
			DockerImageReference: reference,
		},
	}
}

func setup(t *testing.T) (*client.Client, *client.Client, func()) {
	masterConfig, clusterAdminKubeConfigFile, err := testserver.StartTestMaster()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	clusterAdminClient, err := testutil.GetClusterAdminClient(clusterAdminKubeConfigFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	clusterAdminKubeConfig, err := testutil.GetClusterAdminClientConfig(clusterAdminKubeConfigFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	projectAdminClient, err := testserver.CreateNewProject(clusterAdminClient, *clusterAdminKubeConfig, testutil.Namespace(), testutil.Namespace())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return projectAdminClient, clusterAdminClient, func() {
		testserver.CleanupMasterEtcd(t, masterConfig)
	}
}

func runTest(t *testing.T, testname string, projectAdminClient *client.Client, imageStream *imageapi.ImageStream, imageStreamMapping *imageapi.ImageStreamMapping, config *buildapi.BuildConfig, tag string) {
	created, err := projectAdminClient.BuildConfigs(testutil.Namespace()).Create(config)
	if err != nil {
		t.Fatalf("Couldn't create BuildConfig: %v", err)
	}

	buildWatch, err := projectAdminClient.Builds(testutil.Namespace()).Watch(metav1.ListOptions{ResourceVersion: created.ResourceVersion})
	if err != nil {
		t.Fatalf("Couldn't subscribe to Builds %v", err)
	}
	defer buildWatch.Stop()

	buildConfigWatch, err := projectAdminClient.BuildConfigs(testutil.Namespace()).Watch(metav1.ListOptions{ResourceVersion: created.ResourceVersion})
	if err != nil {
		t.Fatalf("Couldn't subscribe to BuildConfigs %v", err)
	}
	defer buildConfigWatch.Stop()

	imageStream, err = projectAdminClient.ImageStreams(testutil.Namespace()).Create(imageStream)
	if err != nil {
		t.Fatalf("Couldn't create ImageStream: %v", err)
	}

	err = projectAdminClient.ImageStreamMappings(testutil.Namespace()).Create(imageStreamMapping)
	if err != nil {
		t.Fatalf("Couldn't create Image: %v", err)
	}

	// wait for initial build event from the creation of the imagerepo with tag latest
	event := <-buildWatch.ResultChan()
	if e, a := watchapi.Added, event.Type; e != a {
		t.Fatalf("expected watch event type %s, got %s", e, a)
	}
	newBuild := event.Object.(*buildapi.Build)
	build1Name := newBuild.Name
	strategy := newBuild.Spec.Strategy
	switch {
	case strategy.SourceStrategy != nil:
		if strategy.SourceStrategy.From.Name != "registry:8080/openshift/test-image-trigger:"+tag {
			i, _ := projectAdminClient.ImageStreams(testutil.Namespace()).Get(imageStream.Name, metav1.GetOptions{})
			bc, _ := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
			t.Fatalf("Expected build with base image %s, got %s\n, imagerepo is %v\ntrigger is %s\n", "registry:8080/openshift/test-image-trigger:"+tag, strategy.SourceStrategy.From.Name, i, bc.Spec.Triggers[0].ImageChange)
		}
	case strategy.DockerStrategy != nil:
		if strategy.DockerStrategy.From.Name != "registry:8080/openshift/test-image-trigger:"+tag {
			i, _ := projectAdminClient.ImageStreams(testutil.Namespace()).Get(imageStream.Name, metav1.GetOptions{})
			bc, _ := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
			t.Fatalf("Expected build with base image %s, got %s\n, imagerepo is %v\ntrigger is %s\n", "registry:8080/openshift/test-image-trigger:"+tag, strategy.DockerStrategy.From.Name, i, bc.Spec.Triggers[0].ImageChange)
		}
	case strategy.CustomStrategy != nil:
		if strategy.CustomStrategy.From.Name != "registry:8080/openshift/test-image-trigger:"+tag {
			i, _ := projectAdminClient.ImageStreams(testutil.Namespace()).Get(imageStream.Name, metav1.GetOptions{})
			bc, _ := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
			t.Fatalf("Expected build with base image %s, got %s\n, imagerepo is %v\ntrigger is %s\n", "registry:8080/openshift/test-image-trigger:"+tag, strategy.CustomStrategy.From.Name, i, bc.Spec.Triggers[0].ImageChange)
		}
	}
	event = <-buildWatch.ResultChan()
	if e, a := watchapi.Modified, event.Type; e != a {
		t.Fatalf("expected watch event type %s, got %s: %#v", e, a, event.Object)
	}
	newBuild = event.Object.(*buildapi.Build)
	// Make sure the resolution of the build's docker image pushspec didn't mutate the persisted API object
	if newBuild.Spec.Output.To.Name != "test-image-trigger-repo:outputtag" {
		t.Fatalf("unexpected build output: %#v %#v", newBuild.Spec.Output.To, newBuild.Spec.Output)
	}
	if newBuild.Labels["testlabel"] != "testvalue" {
		t.Fatalf("Expected build with label %s=%s from build config got %s=%s", "testlabel", "testvalue", "testlabel", newBuild.Labels["testlabel"])
	}

	// wait for build config to be updated
	<-buildConfigWatch.ResultChan()
	updatedConfig, err := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Couldn't get BuildConfig: %v", err)
	}
	// the first tag did not have an image id, so the last trigger field is the pull spec
	if updatedConfig.Spec.Triggers[0].ImageChange.LastTriggeredImageID != "registry:8080/openshift/test-image-trigger:"+tag {
		t.Errorf("Expected imageID equal to pull spec, got %#v", updatedConfig.Spec.Triggers[0].ImageChange)
	}

	// trigger a build by posting a new image
	if err := projectAdminClient.ImageStreamMappings(testutil.Namespace()).Create(&imageapi.ImageStreamMapping{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testutil.Namespace(),
			Name:      imageStream.Name,
		},
		Tag: tag,
		Image: imageapi.Image{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ref-2-random",
			},
			DockerImageReference: "registry:8080/openshift/test-image-trigger:ref-2-random",
		},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// throw away events from build1, we only care about the new build
	// we just triggered
	for {
		event = <-buildWatch.ResultChan()
		newBuild = event.Object.(*buildapi.Build)
		if newBuild.Name != build1Name {
			break
		}
	}
	if e, a := watchapi.Added, event.Type; e != a {
		t.Fatalf("expected watch event type %s, got %s", e, a)
	}
	strategy = newBuild.Spec.Strategy
	switch {
	case strategy.SourceStrategy != nil:
		if strategy.SourceStrategy.From.Name != "registry:8080/openshift/test-image-trigger:ref-2-random" {
			i, _ := projectAdminClient.ImageStreams(testutil.Namespace()).Get(imageStream.Name, metav1.GetOptions{})
			bc, _ := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
			t.Fatalf("Expected build with base image %s, got %s\n, imagerepo is %v\trigger is %s\n", "registry:8080/openshift/test-image-trigger:ref-2-random", strategy.SourceStrategy.From.Name, i, bc.Spec.Triggers[3].ImageChange)
		}
	case strategy.DockerStrategy != nil:
		if strategy.DockerStrategy.From.Name != "registry:8080/openshift/test-image-trigger:ref-2-random" {
			i, _ := projectAdminClient.ImageStreams(testutil.Namespace()).Get(imageStream.Name, metav1.GetOptions{})
			bc, _ := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
			t.Fatalf("Expected build with base image %s, got %s\n, imagerepo is %v\trigger is %s\n", "registry:8080/openshift/test-image-trigger:ref-2-random", strategy.DockerStrategy.From.Name, i, bc.Spec.Triggers[3].ImageChange)
		}
	case strategy.CustomStrategy != nil:
		if strategy.CustomStrategy.From.Name != "registry:8080/openshift/test-image-trigger:ref-2-random" {
			i, _ := projectAdminClient.ImageStreams(testutil.Namespace()).Get(imageStream.Name, metav1.GetOptions{})
			bc, _ := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
			t.Fatalf("Expected build with base image %s, got %s\n, imagerepo is %v\trigger is %s\n", "registry:8080/openshift/test-image-trigger:ref-2-random", strategy.CustomStrategy.From.Name, i, bc.Spec.Triggers[3].ImageChange)
		}
	}

	// throw away events from build1, we only care about the new build
	// we just triggered
	for {
		event = <-buildWatch.ResultChan()
		newBuild = event.Object.(*buildapi.Build)
		if newBuild.Name != build1Name {
			break
		}
	}
	if e, a := watchapi.Modified, event.Type; e != a {
		t.Fatalf("expected watch event type %s, got %s", e, a)
	}
	// Make sure the resolution of the build's docker image pushspec didn't mutate the persisted API object
	if newBuild.Spec.Output.To.Name != "test-image-trigger-repo:outputtag" {
		t.Fatalf("unexpected build output: %#v %#v", newBuild.Spec.Output.To, newBuild.Spec.Output)
	}
	if newBuild.Labels["testlabel"] != "testvalue" {
		t.Fatalf("Expected build with label %s=%s from build config got %s=%s", "testlabel", "testvalue", "testlabel", newBuild.Labels["testlabel"])
	}

	<-buildConfigWatch.ResultChan()
	updatedConfig, err = projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Couldn't get BuildConfig: %v", err)
	}
	if e, a := "registry:8080/openshift/test-image-trigger:ref-2-random", updatedConfig.Spec.Triggers[0].ImageChange.LastTriggeredImageID; e != a {
		t.Errorf("unexpected trigger id: expected %v, got %v", e, a)
	}
}

func TestMultipleImageChangeBuildTriggers(t *testing.T) {
	mockImageStream := func(name, tag string) *imageapi.ImageStream {
		return &imageapi.ImageStream{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: imageapi.ImageStreamSpec{
				DockerImageRepository: "registry:5000/openshift/" + name,
				Tags: map[string]imageapi.TagReference{
					tag: {
						From: &kapi.ObjectReference{
							Kind: "DockerImage",
							Name: "registry:5000/openshift/" + name + ":" + tag,
						},
					},
				},
			},
		}

	}
	mockStreamMapping := func(name, tag string) *imageapi.ImageStreamMapping {
		return &imageapi.ImageStreamMapping{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Tag:        tag,
			Image: imageapi.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				DockerImageReference: "registry:5000/openshift/" + name + ":" + tag,
			},
		}

	}
	multipleImageChangeBuildConfig := func() *buildapi.BuildConfig {
		strategy := stiStrategy("ImageStreamTag", "image1:tag1")
		bc := imageChangeBuildConfig("multi-image-trigger", strategy)
		bc.Spec.CommonSpec.Output.To.Name = "image1:outputtag"
		bc.Spec.Triggers = []buildapi.BuildTriggerPolicy{
			{
				Type:        buildapi.ImageChangeBuildTriggerType,
				ImageChange: &buildapi.ImageChangeTrigger{},
			},
			{
				Type: buildapi.ImageChangeBuildTriggerType,
				ImageChange: &buildapi.ImageChangeTrigger{
					From: &kapi.ObjectReference{
						Name: "image2:tag2",
						Kind: "ImageStreamTag",
					},
				},
			},
			{
				Type: buildapi.ImageChangeBuildTriggerType,
				ImageChange: &buildapi.ImageChangeTrigger{
					From: &kapi.ObjectReference{
						Name: "image3:tag3",
						Kind: "ImageStreamTag",
					},
				},
			},
		}
		return bc
	}
	projectAdminClient, _, fn := setup(t)
	defer fn()
	config := multipleImageChangeBuildConfig()
	triggersToTest := []struct {
		triggerIndex int
		name         string
		tag          string
	}{
		{
			triggerIndex: 0,
			name:         "image1",
			tag:          "tag1",
		},
		{
			triggerIndex: 1,
			name:         "image2",
			tag:          "tag2",
		},
		{
			triggerIndex: 2,
			name:         "image3",
			tag:          "tag3",
		},
	}

	created, err := projectAdminClient.BuildConfigs(testutil.Namespace()).Create(config)
	if err != nil {
		t.Fatalf("Couldn't create BuildConfig: %v", err)
	}
	buildWatch, err := projectAdminClient.Builds(testutil.Namespace()).Watch(metav1.ListOptions{ResourceVersion: created.ResourceVersion})
	if err != nil {
		t.Fatalf("Couldn't subscribe to Builds %v", err)
	}
	defer buildWatch.Stop()

	buildConfigWatch, err := projectAdminClient.BuildConfigs(testutil.Namespace()).Watch(metav1.ListOptions{ResourceVersion: created.ResourceVersion})
	if err != nil {
		t.Fatalf("Couldn't subscribe to BuildConfigs %v", err)
	}
	defer buildConfigWatch.Stop()

	// Builds can continue to produce new events that we don't care about for this test,
	// so once we've seen the last event we care about for a build, we add it to this
	// list so we can ignore additional events from that build.
	ignoreBuilds := make(map[string]struct{})

	for _, tc := range triggersToTest {
		imageStream := mockImageStream(tc.name, tc.tag)
		imageStreamMapping := mockStreamMapping(tc.name, tc.tag)
		imageStream, err = projectAdminClient.ImageStreams(testutil.Namespace()).Create(imageStream)
		if err != nil {
			t.Fatalf("Couldn't create ImageStream: %v", err)
		}

		err = projectAdminClient.ImageStreamMappings(testutil.Namespace()).Create(imageStreamMapping)
		if err != nil {
			t.Fatalf("Couldn't create Image: %v", err)
		}

		var newBuild *buildapi.Build
		var event watchapi.Event
		// wait for initial build event from the creation of the imagerepo
		newBuild, event = filterEvents(t, ignoreBuilds, buildWatch)
		if e, a := watchapi.Added, event.Type; e != a {
			t.Fatalf("expected watch event type %s, got %s", e, a)
		}

		trigger := config.Spec.Triggers[tc.triggerIndex]
		if trigger.ImageChange.From == nil {
			strategy := newBuild.Spec.Strategy
			switch {
			case strategy.SourceStrategy != nil:
				if strategy.SourceStrategy.From.Name != "registry:5000/openshift/"+tc.name+":"+tc.tag {
					i, _ := projectAdminClient.ImageStreams(testutil.Namespace()).Get(imageStream.Name, metav1.GetOptions{})
					bc, _ := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
					t.Fatalf("Expected build with base image %s, got %s\n, imagerepo is %v\ntrigger is %#v", "registry:5000/openshift/"+tc.name+":"+tc.tag, strategy.SourceStrategy.From.Name, i, bc.Spec.Triggers[tc.triggerIndex].ImageChange)
				}
			case strategy.DockerStrategy != nil:
				if strategy.DockerStrategy.From.Name != "registry:8080/openshift/"+tc.name+":"+tc.tag {
					i, _ := projectAdminClient.ImageStreams(testutil.Namespace()).Get(imageStream.Name, metav1.GetOptions{})
					bc, _ := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
					t.Fatalf("Expected build with base image %s, got %s\n, imagerepo is %v\ntrigger is %#v", "registry:5000/openshift/"+tc.name+":"+tag, strategy.DockerStrategy.From.Name, i, bc.Spec.Triggers[tc.triggerIndex].ImageChange)
				}
			case strategy.CustomStrategy != nil:
				if strategy.CustomStrategy.From.Name != "registry:8080/openshift/"+tc.name+":"+tag {
					i, _ := projectAdminClient.ImageStreams(testutil.Namespace()).Get(imageStream.Name, metav1.GetOptions{})
					bc, _ := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
					t.Fatalf("Expected build with base image %s, got %s\n, imagerepo is %v\ntrigger is %#v", "registry:5000/openshift/"+tc.name+":"+tag, strategy.CustomStrategy.From.Name, i, bc.Spec.Triggers[tc.triggerIndex].ImageChange)
				}

			}
		}
		newBuild, event = filterEvents(t, ignoreBuilds, buildWatch)
		if e, a := watchapi.Modified, event.Type; e != a {
			t.Fatalf("expected watch event type %s, got %s", e, a)
		}
		// Make sure the resolution of the build's docker image pushspec didn't mutate the persisted API object
		if newBuild.Spec.Output.To.Name != "image1:outputtag" {
			t.Fatalf("unexpected build output: %#v %#v", newBuild.Spec.Output.To, newBuild.Spec.Output)
		}

		// wait for build config to be updated
		<-buildConfigWatch.ResultChan()
		updatedConfig, err := projectAdminClient.BuildConfigs(testutil.Namespace()).Get(config.Name, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Couldn't get BuildConfig: %v", err)
		}
		// the first tag did not have an image id, so the last trigger field is the pull spec
		if updatedConfig.Spec.Triggers[tc.triggerIndex].ImageChange.LastTriggeredImageID != "registry:5000/openshift/"+tc.name+":"+tc.tag {
			t.Fatalf("Expected imageID equal to pull spec, got %#v", updatedConfig.Spec.Triggers[0].ImageChange)
		}

		ignoreBuilds[newBuild.Name] = struct{}{}

	}
}

func filterEvents(t *testing.T, ignoreBuilds map[string]struct{}, buildWatch watchapi.Interface) (newBuild *buildapi.Build, event watchapi.Event) {
	for {
		event = <-buildWatch.ResultChan()
		var ok bool
		newBuild, ok = event.Object.(*buildapi.Build)
		if !ok {
			t.Errorf("unexpected event type (not a Build): %v", event.Object)
		}
		if _, exists := ignoreBuilds[newBuild.Name]; !exists {
			break
		}
	}
	return
}
