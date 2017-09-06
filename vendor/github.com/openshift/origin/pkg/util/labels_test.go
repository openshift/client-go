package util

import (
	"reflect"
	"testing"

	kmeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kapi "k8s.io/kubernetes/pkg/api"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
)

type FakeLabelsResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func (obj *FakeLabelsResource) GetObjectKind() schema.ObjectKind { return &obj.TypeMeta }

func TestAddConfigLabels(t *testing.T) {
	var nilLabels map[string]string

	testCases := []struct {
		obj            runtime.Object
		addLabels      map[string]string
		err            bool
		expectedLabels map[string]string
	}{
		{ // [0] Test nil + nil => nil
			obj:            &kapi.Pod{},
			addLabels:      nilLabels,
			err:            false,
			expectedLabels: nilLabels,
		},
		{ // [1] Test nil + empty labels => empty labels
			obj:            &kapi.Pod{},
			addLabels:      map[string]string{},
			err:            false,
			expectedLabels: map[string]string{},
		},
		{ // [2] Test obj.Labels + nil => obj.Labels
			obj: &kapi.Pod{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"foo": "bar"}},
			},
			addLabels:      nilLabels,
			err:            false,
			expectedLabels: map[string]string{"foo": "bar"},
		},
		{ // [3] Test obj.Labels + empty labels => obj.Labels
			obj: &kapi.Pod{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"foo": "bar"}},
			},
			addLabels:      map[string]string{},
			err:            false,
			expectedLabels: map[string]string{"foo": "bar"},
		},
		{ // [4] Test nil + addLabels => addLabels
			obj:            &kapi.Pod{},
			addLabels:      map[string]string{"foo": "bar"},
			err:            false,
			expectedLabels: map[string]string{"foo": "bar"},
		},
		{ // [5] Test obj.labels + addLabels => expectedLabels
			obj: &kapi.Service{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"baz": ""}},
			},
			addLabels:      map[string]string{"foo": "bar"},
			err:            false,
			expectedLabels: map[string]string{"foo": "bar", "baz": ""},
		},
		{ // [6] Test conflicting keys with the same value
			obj: &kapi.Service{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"foo": "same value"}},
			},
			addLabels:      map[string]string{"foo": "same value"},
			err:            false,
			expectedLabels: map[string]string{"foo": "same value"},
		},
		{ // [7] Test conflicting keys with a different value
			obj: &kapi.Service{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"foo": "first value"}},
			},
			addLabels:      map[string]string{"foo": "second value"},
			err:            false,
			expectedLabels: map[string]string{"foo": "second value"},
		},
		{ // [8] Test conflicting keys with the same value in ReplicationController nested labels
			obj: &kapi.ReplicationController{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"foo": "same value"},
				},
				Spec: kapi.ReplicationControllerSpec{
					Template: &kapi.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{},
						},
					},
				},
			},
			addLabels:      map[string]string{"foo": "same value"},
			err:            false,
			expectedLabels: map[string]string{"foo": "same value"},
		},
		{ // [9] Test adding labels to a DeploymentConfig object
			obj: &deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"foo": "first value"},
				},
				Spec: deployapi.DeploymentConfigSpec{
					Template: &kapi.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"foo": "first value"},
						},
					},
				},
			},
			addLabels:      map[string]string{"bar": "second value"},
			err:            false,
			expectedLabels: map[string]string{"foo": "first value", "bar": "second value"},
		},
		{ // [10] Test unknown Generic Object with Labels field
			obj: &FakeLabelsResource{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"baz": ""}},
			},
			addLabels:      map[string]string{"foo": "bar"},
			err:            false,
			expectedLabels: map[string]string{"foo": "bar", "baz": ""},
		},
	}

	for i, test := range testCases {
		err := AddObjectLabels(test.obj, test.addLabels)
		if err != nil && !test.err {
			t.Errorf("Unexpected error while setting labels on testCase[%v]: %v.", i, err)
		} else if err == nil && test.err {
			t.Errorf("Unexpected non-error while setting labels on testCase[%v].", i)
		}

		accessor, err := kmeta.Accessor(test.obj)
		if err != nil {
			t.Error(err)
		}
		metaLabels := accessor.GetLabels()
		if e, a := test.expectedLabels, metaLabels; !reflect.DeepEqual(e, a) {
			t.Errorf("Unexpected labels on testCase[%v]. Expected: %#v, got: %#v.", i, e, a)
		}

		// must not add any new nested labels
		switch objType := test.obj.(type) {
		case *kapi.ReplicationController:
			if e, a := map[string]string{}, objType.Spec.Template.Labels; !reflect.DeepEqual(e, a) {
				t.Errorf("Unexpected labels on testCase[%v]. Expected: %#v, got: %#v.", i, e, a)
			}
		case *deployapi.DeploymentConfig:
			if e, a := test.expectedLabels, objType.Spec.Template.Labels; !reflect.DeepEqual(e, a) {
				t.Errorf("Unexpected labels on testCase[%v]. Expected: %#v, got: %#v.", i, e, a)
			}
		}
	}
}

func TestMergeInto(t *testing.T) {
	var nilMap map[int]int

	testCases := []struct {
		dst      interface{}
		src      interface{}
		flags    int
		err      bool
		expected interface{}
	}{
		{ // [0] Can't merge into nil
			dst:      nil,
			src:      map[int]int{},
			flags:    0,
			err:      true,
			expected: nil,
		},
		{ // [1] Can't merge untyped nil into an empty map
			dst:      map[int]int{},
			src:      nil,
			flags:    0,
			err:      true,
			expected: map[int]int{},
		},
		{ // [2] Merge nil map into an empty map
			dst:      map[int]int{},
			src:      nilMap,
			flags:    0,
			err:      false,
			expected: map[int]int{},
		},
		{ // [3] Can't merge into nil map
			dst:      nilMap,
			src:      map[int]int{},
			flags:    0,
			err:      true,
			expected: nilMap,
		},
		{ // [4] Can't merge into pointer
			dst:      &nilMap,
			src:      map[int]int{},
			flags:    0,
			err:      true,
			expected: &nilMap,
		},
		{ // [5] Test empty maps
			dst:      map[int]int{},
			src:      map[int]int{},
			flags:    0,
			err:      false,
			expected: map[int]int{},
		},
		{ // [6] Test dst + src => expected
			dst:      map[int]byte{0: 0, 1: 1},
			src:      map[int]byte{2: 2, 3: 3},
			flags:    0,
			err:      false,
			expected: map[int]byte{0: 0, 1: 1, 2: 2, 3: 3},
		},
		{ // [7] Test dst + src => expected, do not overwrite dst
			dst:      map[string]string{"foo": "bar"},
			src:      map[string]string{"foo": ""},
			flags:    0,
			err:      false,
			expected: map[string]string{"foo": "bar"},
		},
		{ // [8] Test dst + src => expected, overwrite dst
			dst:      map[string]string{"foo": "bar"},
			src:      map[string]string{"foo": ""},
			flags:    OverwriteExistingDstKey,
			err:      false,
			expected: map[string]string{"foo": ""},
		},
		{ // [9] Test dst + src => expected, error on existing key value
			dst:      map[string]string{"foo": "bar"},
			src:      map[string]string{"foo": "bar"},
			flags:    ErrorOnExistingDstKey | OverwriteExistingDstKey,
			err:      true,
			expected: map[string]string{"foo": "bar"},
		},
		{ // [10] Test dst + src => expected, do not error on same key value
			dst:      map[string]string{"foo": "bar"},
			src:      map[string]string{"foo": "bar"},
			flags:    ErrorOnDifferentDstKeyValue | OverwriteExistingDstKey,
			err:      false,
			expected: map[string]string{"foo": "bar"},
		},
		{ // [11] Test dst + src => expected, error on different key value
			dst:      map[string]string{"foo": "bar"},
			src:      map[string]string{"foo": ""},
			flags:    ErrorOnDifferentDstKeyValue | OverwriteExistingDstKey,
			err:      true,
			expected: map[string]string{"foo": "bar"},
		},
	}

	for i, test := range testCases {
		err := MergeInto(test.dst, test.src, test.flags)
		if err != nil && !test.err {
			t.Errorf("Unexpected error while merging maps on testCase[%v]: %v.", i, err)
		} else if err == nil && test.err {
			t.Errorf("Unexpected non-error while merging maps on testCase[%v].", i)
		}

		if !reflect.DeepEqual(test.dst, test.expected) {
			t.Errorf("Unexpected map on testCase[%v]. Expected: %#v, got: %#v.", i, test.expected, test.dst)
		}
	}
}
