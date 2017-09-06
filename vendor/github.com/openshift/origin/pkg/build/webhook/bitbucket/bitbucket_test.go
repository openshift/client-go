package bitbucket

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	kapi "k8s.io/kubernetes/pkg/api"

	buildapi "github.com/openshift/origin/pkg/build/apis/build"
	"github.com/openshift/origin/pkg/build/webhook"
)

var testBuildConfig = &buildapi.BuildConfig{
	Spec: buildapi.BuildConfigSpec{
		Triggers: []buildapi.BuildTriggerPolicy{
			{
				Type: buildapi.BitbucketWebHookBuildTriggerType,
				BitbucketWebHook: &buildapi.WebHookTrigger{
					Secret: "secret100",
				},
			},
		},
		CommonSpec: buildapi.CommonSpec{
			Source: buildapi.BuildSource{
				Git: &buildapi.GitBuildSource{
					URI: "git://bitbucket.com/my/repo.git",
				},
			},
			Strategy: mockBuildStrategy,
		},
	},
}

var mockBuildStrategy = buildapi.BuildStrategy{
	SourceStrategy: &buildapi.SourceBuildStrategy{
		From: kapi.ObjectReference{
			Kind: "DockerImage",
			Name: "repository/image",
		},
	},
}

type okBuildConfigInstantiator struct{}

func (*okBuildConfigInstantiator) Instantiate(namespace string, request *buildapi.BuildRequest) (*buildapi.Build, error) {
	return &buildapi.Build{}, nil
}

type fakeResponder struct {
	called     bool
	statusCode int
	object     runtime.Object
	err        error
}

func (r *fakeResponder) Object(statusCode int, obj runtime.Object) {
	if r.called {
		panic("called twice")
	}
	r.called = true
	r.statusCode = statusCode
	r.object = obj
}

func (r *fakeResponder) Error(err error) {
	if r.called {
		panic("called twice")
	}
	r.called = true
	r.err = err
}

var buildConfig = &buildapi.BuildConfig{
	Spec: buildapi.BuildConfigSpec{
		Triggers: []buildapi.BuildTriggerPolicy{
			{
				Type: buildapi.BitbucketWebHookBuildTriggerType,
				BitbucketWebHook: &buildapi.WebHookTrigger{
					Secret: "secret100",
				},
			},
		},
		CommonSpec: buildapi.CommonSpec{
			Source: buildapi.BuildSource{
				Git: &buildapi.GitBuildSource{},
			},
		},
	},
}

func GivenRequest(method string) *http.Request {
	req, _ := http.NewRequest(method, "http://someurl.com", nil)
	return req
}

func TestVerifyRequestForMethod(t *testing.T) {
	req := GivenRequest("GET")
	plugin := New()
	revision, _, _, proceed, err := plugin.Extract(buildConfig, "secret100", "", req)

	if err == nil || !strings.Contains(err.Error(), "unsupported HTTP method") {
		t.Errorf("Expected unsupported HTTP method, got %v", err)
	}
	if proceed {
		t.Error("Expected 'proceed' return value to be 'false'")
	}
	if revision != nil {
		t.Error("Expected the 'revision' return value to be nil")
	}
}

func TestWrongSecret(t *testing.T) {
	req := GivenRequest("POST")
	plugin := New()
	revision, _, _, proceed, err := plugin.Extract(buildConfig, "wrongsecret", "", req)

	if err != webhook.ErrSecretMismatch {
		t.Errorf("Expected %v, got %v", webhook.ErrSecretMismatch, err)
	}
	if proceed {
		t.Error("Expected 'proceed' return value to be 'false'")
	}
	if revision != nil {
		t.Error("Expected the 'revision' return value to be nil")
	}
}

func TestMissingEvent(t *testing.T) {
	req := GivenRequest("POST")
	req.Header.Add("Content-Type", "application/json")
	plugin := New()
	revision, _, _, proceed, err := plugin.Extract(buildConfig, "secret100", "", req)

	if err == nil || !strings.Contains(err.Error(), "missing X-Event-Key") {
		t.Errorf("Expected missing X-Event-Key, got %v", err)
	}
	if proceed {
		t.Error("Expected 'proceed' return value to be 'false'")
	}
	if revision != nil {
		t.Error("Expected the 'revision' return value to be nil")
	}
}

func TestWrongEventKey(t *testing.T) {
	req := GivenRequest("POST")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Event-Key", "wrong")
	plugin := New()
	revision, _, _, proceed, err := plugin.Extract(buildConfig, "secret100", "", req)

	if err == nil || !strings.Contains(err.Error(), "Unknown Bitbucket X-Event-Key") {
		t.Errorf("Expected missing Unknown Bitbucket X-Event-Key, got %v", err)
	}
	if proceed {
		t.Error("Expected 'proceed' return value to be 'false'")
	}
	if revision != nil {
		t.Error("Expected the 'revision' return value to be nil")
	}
}

func TestJsonPushEventError(t *testing.T) {
	req := post("X-Event-Key", "repo:push", []byte{}, "http://some.url", http.StatusBadRequest, t)
	plugin := New()
	revision, _, _, proceed, err := plugin.Extract(buildConfig, "secret100", "", req)

	if err == nil || !strings.Contains(err.Error(), "unexpected end of JSON input") {
		t.Errorf("Expected unexpected end of JSON input, got %v", err)
	}
	if proceed {
		t.Error("Expected 'proceed' return value to be 'false'")
	}
	if revision != nil {
		t.Error("Expected the 'revision' return value to be nil")
	}
}

func TestJsonBitbucketPushEvent(t *testing.T) {
	req := postFile("X-Event-Key", "repo:push", "pushevent.json", "http://some.url", http.StatusOK, t)
	plugin := New()
	_, _, _, proceed, err := plugin.Extract(buildConfig, "secret100", "", req)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !proceed {
		t.Error("Expected 'proceed' return value to be 'true'")
	}
}

func TestJsonGitHubPushEventWithCharset(t *testing.T) {
	req := postFileWithCharset("X-Event-Key", "repo:push", "pushevent.json", "http://some.url", "application/json; charset=utf-8", http.StatusOK, t)
	plugin := New()
	_, _, _, proceed, err := plugin.Extract(buildConfig, "secret100", "", req)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !proceed {
		t.Error("Expected 'proceed' return value to be 'true'")
	}
}

func postFile(eventHeader, eventName, filename, url string, expStatusCode int, t *testing.T) *http.Request {
	return postFileWithCharset(eventHeader, eventName, filename, url, "application/json", expStatusCode, t)
}

func postFileWithCharset(eventHeader, eventName, filename, url, charset string, expStatusCode int, t *testing.T) *http.Request {
	data, err := ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		t.Errorf("Failed to open %s: %v", filename, err)
	}

	return postWithCharset(eventHeader, eventName, data, url, charset, expStatusCode, t)
}

func post(eventHeader, eventName string, data []byte, url string, expStatusCode int, t *testing.T) *http.Request {
	return postWithCharset(eventHeader, eventName, data, url, "application/json", expStatusCode, t)
}

func postWithCharset(eventHeader, eventName string, data []byte, url, charset string, expStatusCode int, t *testing.T) *http.Request {
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		t.Errorf("Error creating POST request: %v", err)
	}

	req.Header.Add("Content-Type", charset)
	req.Header.Add(eventHeader, eventName)

	return req
}

type testContext struct {
	plugin   WebHook
	buildCfg *buildapi.BuildConfig
	req      *http.Request
	path     string
}

func setup(t *testing.T, filename, eventType, ref string) *testContext {
	context := testContext{
		plugin: WebHook{},
		buildCfg: &buildapi.BuildConfig{
			Spec: buildapi.BuildConfigSpec{
				Triggers: []buildapi.BuildTriggerPolicy{
					{
						Type: buildapi.BitbucketWebHookBuildTriggerType,
						BitbucketWebHook: &buildapi.WebHookTrigger{
							Secret: "secret100",
						},
					},
				},
				CommonSpec: buildapi.CommonSpec{
					Source: buildapi.BuildSource{
						Git: &buildapi.GitBuildSource{
							URI: "git://bitbucket.com/my/repo.git",
							Ref: ref,
						},
					},
					Strategy: mockBuildStrategy,
				},
			},
		},
		path: "/foobar",
	}
	event, err := ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		t.Errorf("Failed to open %s: %v", filename, err)
	}
	req, err := http.NewRequest("POST", "http://origin.com", bytes.NewReader(event))
	if err != nil {
		t.Errorf("Failed to create a new request (%s)", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Event-Key", eventType)

	context.req = req
	return &context
}

func TestExtractProvidesValidBuildForAPushEvent(t *testing.T) {
	//setup
	context := setup(t, "pushevent.json", "repo:push", "")

	//execute
	revision, _, _, proceed, err := context.plugin.Extract(context.buildCfg, "secret100", context.path, context.req)

	//validation
	if err != nil {
		t.Errorf("Error while extracting build info: %s", err)
	}
	if !proceed {
		t.Errorf("The 'proceed' return value should equal 'true' %t", proceed)
	}
	if revision == nil {
		t.Fatal("Expecting the revision to not be nil")
	}
	if revision.Git.Commit != "03f4a7270240708834de475bcf21532d6134777e" {
		t.Error("Expecting the revision to contain the commit id from the push event")
	}

}

func TestExtractProvidesValidBuildForAPushEventOtherThanMaster(t *testing.T) {
	//setup
	context := setup(t, "pushevent-not-master.json", "repo:push", "this-is-not-master")
	//execute
	revision, _, _, proceed, err := context.plugin.Extract(context.buildCfg, "secret100", context.path, context.req)

	//validation
	if err != nil {
		t.Errorf("Error while extracting build info: %s", err)
	}
	if !proceed {
		t.Errorf("The 'proceed' return value should equal 'true' %t", proceed)
	}
	if revision == nil {
		t.Fatal("Expecting the revision to not be nil")
	}
	if revision.Git.Commit != "03f4a7270240708834de475bcf21532d6134777e" {
		t.Error("Expecting the revision to contain the commit id from the push event")
	}
}

func TestExtractSkipsBuildForUnmatchedBranches(t *testing.T) {
	//setup
	context := setup(t, "pushevent.json", "push", "wrongref")

	//execute
	_, _, _, proceed, _ := context.plugin.Extract(context.buildCfg, "secret100", context.path, context.req)
	if proceed {
		t.Errorf("Expecting to not continue from this event because the branch is not for this buildConfig '%s'", context.buildCfg.Spec.Source.Git.Ref)
	}
}
