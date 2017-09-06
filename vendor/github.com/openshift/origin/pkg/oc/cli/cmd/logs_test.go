package cmd

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	kcmd "k8s.io/kubernetes/pkg/kubectl/cmd"

	buildapi "github.com/openshift/origin/pkg/build/apis/build"
	buildclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
)

// TestLogsFlagParity makes sure that our copied flags don't slip during rebases
func TestLogsFlagParity(t *testing.T) {
	kubeCmd := kcmd.NewCmdLogs(nil, ioutil.Discard)
	f := clientcmd.NewFactory(nil)
	originCmd := NewCmdLogs("oc", "logs", f, ioutil.Discard)

	kubeCmd.LocalFlags().VisitAll(func(kubeFlag *pflag.Flag) {
		originFlag := originCmd.LocalFlags().Lookup(kubeFlag.Name)
		if originFlag == nil {
			t.Errorf("missing %v flag", kubeFlag.Name)
			return
		}

		if !reflect.DeepEqual(originFlag, kubeFlag) {
			t.Errorf("flag %v %v does not match %v", kubeFlag.Name, kubeFlag, originFlag)
		}
	})
}

type fakeBuildClient struct {
	build *buildapi.Build
}

func (f *fakeBuildClient) List(opts metav1.ListOptions) (*buildapi.BuildList, error) {
	return nil, nil
}

func (f *fakeBuildClient) Get(names string, opts metav1.GetOptions) (*buildapi.Build, error) {
	return f.build, nil
}

func (f *fakeBuildClient) Create(build *buildapi.Build) (*buildapi.Build, error) {
	return nil, nil
}

func (f *fakeBuildClient) Update(build *buildapi.Build) (*buildapi.Build, error) {
	return nil, nil
}

func (f *fakeBuildClient) Delete(name string) error {
	return nil
}

func (f *fakeBuildClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (f *fakeBuildClient) Clone(request *buildapi.BuildRequest) (*buildapi.Build, error) {
	return nil, nil
}

func (f *fakeBuildClient) UpdateDetails(build *buildapi.Build) (*buildapi.Build, error) {
	return nil, nil
}

func (f *fakeBuildClient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*buildapi.Build, error) {
	return nil, nil
}

type fakeNamespacer struct {
	client buildclient.BuildInterface
}

func (f *fakeNamespacer) Builds(namespace string) buildclient.BuildInterface {
	return f.client
}

type fakeWriter struct {
	data []byte
}

func (f *fakeWriter) Write(p []byte) (n int, err error) {
	f.data = p
	return len(p), nil
}

func TestRunLogForPipelineStrategy(t *testing.T) {
	bld := buildapi.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "foo",
			Annotations: map[string]string{buildapi.BuildJenkinsBlueOceanLogURLAnnotation: "https://foo"},
		},
		Spec: buildapi.BuildSpec{
			CommonSpec: buildapi.CommonSpec{
				Strategy: buildapi.BuildStrategy{
					JenkinsPipelineStrategy: &buildapi.JenkinsPipelineBuildStrategy{},
				},
			},
		},
	}

	fakebc := fakeBuildClient{
		build: &bld,
	}
	fakenamespacer := fakeNamespacer{
		client: &fakebc,
	}
	fakewriter := fakeWriter{}

	testCases := []struct {
		o runtime.Object
	}{
		{
			o: &bld,
		},
		{
			o: &buildapi.BuildConfig{
				Spec: buildapi.BuildConfigSpec{
					CommonSpec: buildapi.CommonSpec{
						Strategy: buildapi.BuildStrategy{
							JenkinsPipelineStrategy: &buildapi.JenkinsPipelineBuildStrategy{},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		opts := OpenShiftLogsOptions{
			KubeLogOptions: &kcmd.LogsOptions{
				Object: tc.o,
				Out:    &fakewriter,
			},
			Client: &fakenamespacer,
		}
		err := opts.RunLog()
		if err != nil {
			t.Errorf("%#v: RunLog error %v", tc.o, err)
		}
		output := string(fakewriter.data[:])
		if !strings.Contains(output, "https://foo") {
			t.Errorf("%#v: RunLog did not have https://foo, but rather had: %s", tc.o, output)
		}
	}

}

func TestIsPipelineBuild(t *testing.T) {
	testCases := []struct {
		o          runtime.Object
		isPipeline bool
	}{
		{
			o: &buildapi.Build{
				Spec: buildapi.BuildSpec{
					CommonSpec: buildapi.CommonSpec{
						Strategy: buildapi.BuildStrategy{
							JenkinsPipelineStrategy: &buildapi.JenkinsPipelineBuildStrategy{},
						},
					},
				},
			},
			isPipeline: true,
		},
		{
			o: &buildapi.Build{
				Spec: buildapi.BuildSpec{
					CommonSpec: buildapi.CommonSpec{
						Strategy: buildapi.BuildStrategy{
							SourceStrategy: &buildapi.SourceBuildStrategy{},
						},
					},
				},
			},
			isPipeline: false,
		},
		{
			o: &buildapi.BuildConfig{
				Spec: buildapi.BuildConfigSpec{
					CommonSpec: buildapi.CommonSpec{
						Strategy: buildapi.BuildStrategy{
							JenkinsPipelineStrategy: &buildapi.JenkinsPipelineBuildStrategy{},
						},
					},
				},
			},
			isPipeline: true,
		},
		{
			o: &buildapi.BuildConfig{
				Spec: buildapi.BuildConfigSpec{
					CommonSpec: buildapi.CommonSpec{
						Strategy: buildapi.BuildStrategy{
							DockerStrategy: &buildapi.DockerBuildStrategy{},
						},
					},
				},
			},
			isPipeline: false,
		},
		{
			o:          &deployapi.DeploymentConfig{},
			isPipeline: false,
		},
	}

	for _, tc := range testCases {
		isPipeline, _, _, _, _ := isPipelineBuild(tc.o)
		if isPipeline != tc.isPipeline {
			t.Errorf("%#v, unexpected results expected isPipeline %v returned isPipeline %v", tc.o, tc.isPipeline, isPipeline)
		}
	}
}
