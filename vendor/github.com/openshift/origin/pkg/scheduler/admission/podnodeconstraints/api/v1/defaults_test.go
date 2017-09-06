package v1_test

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"

	configapi "github.com/openshift/origin/pkg/cmd/server/api"
	v1 "github.com/openshift/origin/pkg/cmd/server/api/v1"
	_ "github.com/openshift/origin/pkg/scheduler/admission/podnodeconstraints/api/install"
	podnodeconstraintsv1 "github.com/openshift/origin/pkg/scheduler/admission/podnodeconstraints/api/v1"
)

func roundTrip(t *testing.T, obj runtime.Object) runtime.Object {
	data, err := runtime.Encode(configapi.Codecs.LegacyCodec(v1.SchemeGroupVersion), obj)
	if err != nil {
		t.Errorf("%v\n %#v", err, obj)
		return nil
	}
	obj2, err := runtime.Decode(configapi.Codecs.UniversalDecoder(), data)
	if err != nil {
		t.Errorf("%v\nData: %s\nSource: %#v", err, string(data), obj)
		return nil
	}
	obj3 := reflect.New(reflect.TypeOf(obj).Elem()).Interface().(runtime.Object)
	err = configapi.Scheme.Convert(obj2, obj3, nil)
	if err != nil {
		t.Errorf("%v\nSource: %#v", err, obj2)
		return nil
	}
	return obj3
}

func TestDefaults(t *testing.T) {
	tests := []struct {
		original *podnodeconstraintsv1.PodNodeConstraintsConfig
		expected *podnodeconstraintsv1.PodNodeConstraintsConfig
	}{
		{
			original: &podnodeconstraintsv1.PodNodeConstraintsConfig{},
			expected: &podnodeconstraintsv1.PodNodeConstraintsConfig{
				NodeSelectorLabelBlacklist: []string{"kubernetes.io/hostname"},
			},
		},
	}
	for i, test := range tests {
		t.Logf("test %d", i)
		original := test.original
		expected := test.expected
		obj2 := roundTrip(t, runtime.Object(original))
		got, ok := obj2.(*podnodeconstraintsv1.PodNodeConstraintsConfig)
		if !ok {
			t.Errorf("unexpected object: %v", got)
			t.FailNow()
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got different than expected:\nA:\t%#v\nB:\t%#v\n\nDiff:\n%s\n\n%s", got, expected, diff.ObjectDiff(expected, got), diff.ObjectGoPrintSideBySide(expected, got))
		}
	}
}
