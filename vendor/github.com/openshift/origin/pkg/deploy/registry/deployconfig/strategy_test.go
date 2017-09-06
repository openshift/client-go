package deployconfig

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
	deploytest "github.com/openshift/origin/pkg/deploy/apis/apps/test"
)

var (
	nonDefaultRevisionHistoryLimit = deployapi.DefaultRevisionHistoryLimit + 42
)

func int32ptr(v int32) *int32 {
	return &v
}

func TestDeploymentConfigStrategy(t *testing.T) {
	ctx := apirequest.NewDefaultContext()
	if !CommonStrategy.NamespaceScoped() {
		t.Errorf("DeploymentConfig is namespace scoped")
	}
	if CommonStrategy.AllowCreateOnUpdate() {
		t.Errorf("DeploymentConfig should not allow create on update")
	}
	deploymentConfig := &deployapi.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
		Spec:       deploytest.OkDeploymentConfigSpec(),
	}
	CommonStrategy.PrepareForCreate(ctx, deploymentConfig)
	errs := CommonStrategy.Validate(ctx, deploymentConfig)
	if len(errs) != 0 {
		t.Errorf("Unexpected error validating %v", errs)
	}
	updatedDeploymentConfig := &deployapi.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "bar", Namespace: "default", Generation: 1},
		Spec:       deploytest.OkDeploymentConfigSpec(),
	}
	errs = CommonStrategy.ValidateUpdate(ctx, updatedDeploymentConfig, deploymentConfig)
	if len(errs) == 0 {
		t.Errorf("Expected error validating")
	}
	// name must match, and resource version must be provided
	updatedDeploymentConfig.Name = "foo"
	updatedDeploymentConfig.ResourceVersion = "1"
	errs = CommonStrategy.ValidateUpdate(ctx, updatedDeploymentConfig, deploymentConfig)
	if len(errs) != 0 {
		t.Errorf("Unexpected error validating %v", errs)
	}
	invalidDeploymentConfig := &deployapi.DeploymentConfig{}
	errs = CommonStrategy.Validate(ctx, invalidDeploymentConfig)
	if len(errs) == 0 {
		t.Errorf("Expected error validating")
	}
}

// TestPrepareForUpdate exercises various client updates.
func TestPrepareForUpdate(t *testing.T) {
	ctx := apirequest.NewDefaultContext()
	tests := []struct {
		name string

		prev     runtime.Object
		after    runtime.Object
		expected runtime.Object
	}{
		{
			name:     "latestVersion bump",
			prev:     prevDeployment(),
			after:    afterDeploymentVersionBump(),
			expected: expectedAfterVersionBump(),
		},
		{
			name:     "spec change",
			prev:     prevDeployment(),
			after:    afterDeployment(),
			expected: expectedAfterDeployment(),
		},
	}

	for _, test := range tests {
		strategy{}.PrepareForUpdate(ctx, test.after, test.prev)
		if !reflect.DeepEqual(test.expected, test.after) {
			t.Errorf("%s: unexpected object mismatch! Expected:\n%#v\ngot:\n%#v", test.name, test.expected, test.after)
		}
	}
}

// prevDeployment is the old object tested for both old and new client updates.
func prevDeployment() *deployapi.DeploymentConfig {
	return &deployapi.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default", Generation: 4, Annotations: make(map[string]string)},
		Spec:       deploytest.OkDeploymentConfigSpec(),
		Status:     deploytest.OkDeploymentConfigStatus(1),
	}
}

// afterDeployment is used for a spec change check.
func afterDeployment() *deployapi.DeploymentConfig {
	dc := prevDeployment()
	dc.Spec.Replicas++
	return dc
}

// expectedAfterDeployment is used for a spec change check.
func expectedAfterDeployment() *deployapi.DeploymentConfig {
	dc := afterDeployment()
	dc.Generation++
	return dc
}

// afterDeploymentVersionBump is a deployment config updated to a newer version.
func afterDeploymentVersionBump() *deployapi.DeploymentConfig {
	dc := prevDeployment()
	dc.Status.LatestVersion++
	return dc
}

// expectedAfterVersionBump is the object we expect after a version bump.
func expectedAfterVersionBump() *deployapi.DeploymentConfig {
	dc := afterDeploymentVersionBump()
	dc.Generation++
	return dc
}

func setRevisionHistoryLimit(v *int32, dc *deployapi.DeploymentConfig) *deployapi.DeploymentConfig {
	dc.Spec.RevisionHistoryLimit = v
	return dc
}

func okDeploymentConfig(generation int64) *deployapi.DeploymentConfig {
	dc := deploytest.OkDeploymentConfig(0)
	dc.ObjectMeta.Generation = generation
	return dc
}

func TestLegacyStrategy_PrepareForCreate(t *testing.T) {
	nonDefaultRevisionHistoryLimit := deployapi.DefaultRevisionHistoryLimit + 42
	tt := []struct {
		obj      *deployapi.DeploymentConfig
		expected *deployapi.DeploymentConfig
	}{
		{
			obj: setRevisionHistoryLimit(nil, okDeploymentConfig(0)),
			// Legacy API shall not default RevisionHistoryLimit to maintain backwards compatibility
			expected: setRevisionHistoryLimit(nil, okDeploymentConfig(1)),
		},
		{
			obj:      setRevisionHistoryLimit(&nonDefaultRevisionHistoryLimit, okDeploymentConfig(0)),
			expected: setRevisionHistoryLimit(&nonDefaultRevisionHistoryLimit, okDeploymentConfig(1)),
		},
	}

	ctx := apirequest.NewDefaultContext()

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			LegacyStrategy.PrepareForCreate(ctx, tc.obj)
			if !reflect.DeepEqual(tc.obj, tc.expected) {
				t.Fatalf("LegacyStrategy.PrepareForCreate failed:%s", diff.ObjectReflectDiff(tc.obj, tc.expected))
			}

			errs := LegacyStrategy.Validate(ctx, tc.obj)
			if len(errs) != 0 {
				t.Errorf("Unexpected error validating DeploymentConfig: %v", errs)
			}
		})
	}
}

func TestLegacyStrategy_DefaultGarbageCollectionPolicy(t *testing.T) {
	expected := rest.OrphanDependents
	got := LegacyStrategy.DefaultGarbageCollectionPolicy()
	if got != expected {
		t.Fatalf("Default garbage collection policy for DeploymentConfigs should be %q (not %q)", expected, got)
	}
}

func TestGroupStrategy_PrepareForCreate(t *testing.T) {
	tt := []struct {
		obj      *deployapi.DeploymentConfig
		expected *deployapi.DeploymentConfig
	}{
		{
			obj: setRevisionHistoryLimit(nil, okDeploymentConfig(0)),
			// Group API should default RevisionHistoryLimit
			expected: setRevisionHistoryLimit(int32ptr(deployapi.DefaultRevisionHistoryLimit), okDeploymentConfig(1)),
		},
		{
			obj:      setRevisionHistoryLimit(&nonDefaultRevisionHistoryLimit, okDeploymentConfig(0)),
			expected: setRevisionHistoryLimit(&nonDefaultRevisionHistoryLimit, okDeploymentConfig(1)),
		},
	}

	ctx := apirequest.NewDefaultContext()

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			GroupStrategy.PrepareForCreate(ctx, tc.obj)
			if !reflect.DeepEqual(tc.obj, tc.expected) {
				t.Fatalf("GroupStrategy.PrepareForCreate failed:%s", diff.ObjectReflectDiff(tc.obj, tc.expected))
			}

			errs := GroupStrategy.Validate(ctx, tc.obj)
			if len(errs) != 0 {
				t.Errorf("Unexpected error validating DeploymentConfig: %v", errs)
			}
		})
	}
}
