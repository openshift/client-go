package strategyrestrictions

import (
	"fmt"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/authentication/user"
	clientgotesting "k8s.io/client-go/testing"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	buildapi "github.com/openshift/origin/pkg/build/apis/build"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/client/testclient"
	oadmission "github.com/openshift/origin/pkg/cmd/server/admission"
)

func TestBuildAdmission(t *testing.T) {
	tests := []struct {
		name             string
		kind             schema.GroupKind
		resource         schema.GroupResource
		subResource      string
		object           runtime.Object
		oldObject        runtime.Object
		responseObject   runtime.Object
		reviewResponse   *authorizationapi.SubjectAccessReviewResponse
		expectedResource string
		expectAccept     bool
		expectedError    string
	}{
		{
			name:             "allowed source build",
			object:           testBuild(buildapi.BuildStrategy{SourceStrategy: &buildapi.SourceBuildStrategy{}}),
			kind:             buildapi.Kind("Build"),
			resource:         buildapi.Resource("builds"),
			reviewResponse:   reviewResponse(true, ""),
			expectedResource: authorizationapi.SourceBuildResource,
			expectAccept:     true,
		},
		{
			name:             "allowed source build clone",
			object:           testBuildRequest("buildname"),
			responseObject:   testBuild(buildapi.BuildStrategy{SourceStrategy: &buildapi.SourceBuildStrategy{}}),
			kind:             buildapi.Kind("Build"),
			resource:         buildapi.Resource("builds"),
			subResource:      "clone",
			reviewResponse:   reviewResponse(true, ""),
			expectedResource: authorizationapi.SourceBuildResource,
			expectAccept:     true,
		},
		{
			name:             "denied docker build",
			object:           testBuild(buildapi.BuildStrategy{DockerStrategy: &buildapi.DockerBuildStrategy{}}),
			kind:             buildapi.Kind("Build"),
			resource:         buildapi.Resource("builds"),
			reviewResponse:   reviewResponse(false, "cannot create build of type docker build"),
			expectAccept:     false,
			expectedResource: authorizationapi.DockerBuildResource,
		},
		{
			name:             "denied docker build clone",
			object:           testBuildRequest("buildname"),
			responseObject:   testBuild(buildapi.BuildStrategy{DockerStrategy: &buildapi.DockerBuildStrategy{}}),
			kind:             buildapi.Kind("Build"),
			resource:         buildapi.Resource("builds"),
			subResource:      "clone",
			reviewResponse:   reviewResponse(false, "cannot create build of type docker build"),
			expectAccept:     false,
			expectedResource: authorizationapi.DockerBuildResource,
		},
		{
			name:             "allowed custom build",
			object:           testBuild(buildapi.BuildStrategy{CustomStrategy: &buildapi.CustomBuildStrategy{}}),
			kind:             buildapi.Kind("Build"),
			resource:         buildapi.Resource("builds"),
			reviewResponse:   reviewResponse(true, ""),
			expectedResource: authorizationapi.CustomBuildResource,
			expectAccept:     true,
		},
		{
			name:             "allowed build config",
			object:           testBuildConfig(buildapi.BuildStrategy{DockerStrategy: &buildapi.DockerBuildStrategy{}}),
			kind:             buildapi.Kind("BuildConfig"),
			resource:         buildapi.Resource("buildconfigs"),
			reviewResponse:   reviewResponse(true, ""),
			expectAccept:     true,
			expectedResource: authorizationapi.DockerBuildResource,
		},
		{
			name:             "allowed build config instantiate",
			responseObject:   testBuildConfig(buildapi.BuildStrategy{DockerStrategy: &buildapi.DockerBuildStrategy{}}),
			object:           testBuildRequest("buildname"),
			kind:             buildapi.Kind("Build"),
			resource:         buildapi.Resource("buildconfigs"),
			subResource:      "instantiate",
			reviewResponse:   reviewResponse(true, ""),
			expectAccept:     true,
			expectedResource: authorizationapi.DockerBuildResource,
		},
		{
			name:             "forbidden build config",
			object:           testBuildConfig(buildapi.BuildStrategy{CustomStrategy: &buildapi.CustomBuildStrategy{}}),
			kind:             buildapi.Kind("Build"),
			resource:         buildapi.Resource("buildconfigs"),
			reviewResponse:   reviewResponse(false, ""),
			expectAccept:     false,
			expectedResource: authorizationapi.CustomBuildResource,
		},
		{
			name:             "forbidden build config instantiate",
			responseObject:   testBuildConfig(buildapi.BuildStrategy{CustomStrategy: &buildapi.CustomBuildStrategy{}}),
			object:           testBuildRequest("buildname"),
			kind:             buildapi.Kind("Build"),
			resource:         buildapi.Resource("buildconfigs"),
			subResource:      "instantiate",
			reviewResponse:   reviewResponse(false, ""),
			expectAccept:     false,
			expectedResource: authorizationapi.CustomBuildResource,
		},
		{
			name:           "unrecognized request object",
			object:         &fakeObject{},
			kind:           buildapi.Kind("BuildConfig"),
			resource:       buildapi.Resource("buildconfigs"),
			reviewResponse: reviewResponse(true, ""),
			expectAccept:   false,
			expectedError:  "Internal error occurred: [Unrecognized request object &admission.fakeObject{}, couldn't find ObjectMeta field in admission.fakeObject{}]",
		},
		{
			name:           "details on forbidden docker build",
			object:         testBuild(buildapi.BuildStrategy{DockerStrategy: &buildapi.DockerBuildStrategy{}}),
			kind:           buildapi.Kind("Build"),
			resource:       buildapi.Resource("builds"),
			subResource:    "details",
			reviewResponse: reviewResponse(false, "cannot create build of type docker build"),
			expectAccept:   true,
		},
		{
			name:             "allowed jenkins pipeline build",
			object:           testBuild(buildapi.BuildStrategy{JenkinsPipelineStrategy: &buildapi.JenkinsPipelineBuildStrategy{}}),
			kind:             buildapi.Kind("Build"),
			resource:         buildapi.Resource("builds"),
			reviewResponse:   reviewResponse(true, ""),
			expectedResource: authorizationapi.JenkinsPipelineBuildResource,
			expectAccept:     true,
		},
		{
			name:             "allowed jenkins pipeline build clone",
			object:           testBuildRequest("buildname"),
			responseObject:   testBuild(buildapi.BuildStrategy{JenkinsPipelineStrategy: &buildapi.JenkinsPipelineBuildStrategy{}}),
			kind:             buildapi.Kind("Build"),
			resource:         buildapi.Resource("builds"),
			subResource:      "clone",
			reviewResponse:   reviewResponse(true, ""),
			expectedResource: authorizationapi.JenkinsPipelineBuildResource,
			expectAccept:     true,
		},
	}

	ops := []admission.Operation{admission.Create, admission.Update}
	for _, test := range tests {
		for _, op := range ops {
			client := fakeClient(test.expectedResource, test.reviewResponse, test.responseObject)
			c := NewBuildByStrategy()
			c.(oadmission.WantsOpenshiftClient).SetOpenshiftClient(client)
			attrs := admission.NewAttributesRecord(test.object, test.oldObject, test.kind.WithVersion("version"), "default", "name", test.resource.WithVersion("version"), test.subResource, op, fakeUser())
			err := c.Admit(attrs)
			if err != nil && test.expectAccept {
				t.Errorf("%s: unexpected error: %v", test.name, err)
			}

			if !apierrors.IsForbidden(err) && !test.expectAccept {
				if (len(test.expectedError) != 0) || (test.expectedError == err.Error()) {
					continue
				}
				t.Errorf("%s: expecting reject error, got %v", test.name, err)
			}
		}
	}
}

type fakeObject struct{}

func (*fakeObject) GetObjectKind() schema.ObjectKind { return nil }

func fakeUser() user.Info {
	return &user.DefaultInfo{
		Name: "testuser",
	}
}

func fakeClient(expectedResource string, reviewResponse *authorizationapi.SubjectAccessReviewResponse, obj runtime.Object) client.Interface {
	emptyResponse := &authorizationapi.SubjectAccessReviewResponse{}

	fake := &testclient.Fake{}
	fake.AddReactor("create", "localsubjectaccessreviews", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		review, ok := action.(clientgotesting.CreateAction).GetObject().(*authorizationapi.LocalSubjectAccessReview)
		if !ok {
			return true, emptyResponse, fmt.Errorf("unexpected object received: %#v", review)
		}
		if review.Action.Resource != expectedResource {
			return true, emptyResponse, fmt.Errorf("unexpected resource received: %s. expected: %s",
				review.Action.Resource, expectedResource)
		}
		return true, reviewResponse, nil
	})
	fake.AddReactor("get", "buildconfigs", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, obj, nil
	})
	fake.AddReactor("get", "builds", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, obj, nil
	})

	return fake
}

func testBuild(strategy buildapi.BuildStrategy) *buildapi.Build {
	return &buildapi.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-build",
		},
		Spec: buildapi.BuildSpec{
			CommonSpec: buildapi.CommonSpec{
				Strategy: strategy,
			},
		},
	}
}

func testBuildConfig(strategy buildapi.BuildStrategy) *buildapi.BuildConfig {
	return &buildapi.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-buildconfig",
		},
		Spec: buildapi.BuildConfigSpec{
			CommonSpec: buildapi.CommonSpec{
				Strategy: strategy,
			},
		},
	}
}

func reviewResponse(allowed bool, msg string) *authorizationapi.SubjectAccessReviewResponse {
	return &authorizationapi.SubjectAccessReviewResponse{
		Allowed: allowed,
		Reason:  msg,
	}
}

func testBuildRequest(name string) runtime.Object {
	return &buildapi.BuildRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
