package reaper

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgotesting "k8s.io/client-go/testing"
	kapi "k8s.io/kubernetes/pkg/api"

	"github.com/davecgh/go-spew/spew"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	"github.com/openshift/origin/pkg/client/testclient"
	oauthapi "github.com/openshift/origin/pkg/oauth/apis/oauth"
	securityapi "github.com/openshift/origin/pkg/security/apis/security"
	"github.com/openshift/origin/pkg/security/legacyclient"
	authenticationapi "github.com/openshift/origin/pkg/user/apis/user"
)

var (
	usersResource                     = schema.GroupVersionResource{Group: "", Version: "", Resource: "users"}
	securityContextContraintsResource = schema.GroupVersionResource{Group: "", Version: "", Resource: "securitycontextconstraints"}
	oAuthClientAuthorizationsResource = schema.GroupVersionResource{Group: "", Version: "", Resource: "oauthclientauthorizations"}
)

func TestUserReaper(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		objects  []runtime.Object
		sccs     []runtime.Object
		expected []interface{}
	}{
		{
			name:    "no objects",
			user:    "bob",
			objects: []runtime.Object{},
			expected: []interface{}{
				clientgotesting.DeleteActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "delete", Resource: usersResource}, Name: "bob"},
			},
		},
		{
			name: "cluster bindings",
			user: "bob",
			objects: []runtime.Object{
				&authorizationapi.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{Name: "binding-no-subjects"},
					RoleRef:    kapi.ObjectReference{Name: "role"},
					Subjects:   []kapi.ObjectReference{},
				},
				&authorizationapi.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{Name: "binding-one-subject"},
					RoleRef:    kapi.ObjectReference{Name: "role"},
					Subjects:   []kapi.ObjectReference{{Name: "bob", Kind: "User"}},
				},
				&authorizationapi.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{Name: "binding-mismatched-subject"},
					RoleRef:    kapi.ObjectReference{Name: "role"},
					Subjects:   []kapi.ObjectReference{{Name: "bob"}, {Name: "bob", Kind: "Group"}, {Name: "bob", Kind: "Other"}},
				},
			},
			expected: []interface{}{
				clientgotesting.UpdateActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "update", Resource: clusterRoleBindingsResource}, Object: &authorizationapi.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{Name: "binding-one-subject"},
					RoleRef:    kapi.ObjectReference{Name: "role"},
					Subjects:   []kapi.ObjectReference{},
				}},
				clientgotesting.DeleteActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "delete", Resource: usersResource}, Name: "bob"},
			},
		},
		{
			name: "namespaced bindings",
			user: "bob",
			objects: []runtime.Object{
				&authorizationapi.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{Name: "binding-no-subjects", Namespace: "ns1"},
					RoleRef:    kapi.ObjectReference{Name: "role"},
					Subjects:   []kapi.ObjectReference{},
				},
				&authorizationapi.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{Name: "binding-one-subject", Namespace: "ns2"},
					RoleRef:    kapi.ObjectReference{Name: "role"},
					Subjects:   []kapi.ObjectReference{{Name: "bob", Kind: "User"}},
				},
				&authorizationapi.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{Name: "binding-mismatched-subject", Namespace: "ns3"},
					RoleRef:    kapi.ObjectReference{Name: "role"},
					Subjects:   []kapi.ObjectReference{{Name: "bob"}, {Name: "bob", Kind: "Group"}, {Name: "bob", Kind: "Other"}},
				},
			},
			expected: []interface{}{
				clientgotesting.UpdateActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "update", Resource: roleBindingsResource, Namespace: "ns2"}, Object: &authorizationapi.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{Name: "binding-one-subject", Namespace: "ns2"},
					RoleRef:    kapi.ObjectReference{Name: "role"},
					Subjects:   []kapi.ObjectReference{},
				}},
				clientgotesting.DeleteActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "delete", Resource: usersResource}, Name: "bob"},
			},
		},
		{
			name: "sccs",
			user: "bob",
			sccs: []runtime.Object{
				&securityapi.SecurityContextConstraints{
					ObjectMeta: metav1.ObjectMeta{Name: "scc-no-subjects"},
					Users:      []string{},
				},
				&securityapi.SecurityContextConstraints{
					ObjectMeta: metav1.ObjectMeta{Name: "scc-one-subject"},
					Users:      []string{"bob"},
				},
				&securityapi.SecurityContextConstraints{
					ObjectMeta: metav1.ObjectMeta{Name: "scc-mismatched-subjects"},
					Users:      []string{"bob2"},
					Groups:     []string{"bob"},
				},
			},
			expected: []interface{}{
				clientgotesting.UpdateActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "update", Resource: securityContextContraintsResource}, Object: &securityapi.SecurityContextConstraints{
					ObjectMeta: metav1.ObjectMeta{Name: "scc-one-subject"},
					Users:      []string{},
				}},
				clientgotesting.DeleteActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "delete", Resource: usersResource}, Name: "bob"},
			},
		},
		{
			name: "identities",
			user: "bob",
			objects: []runtime.Object{
				&authenticationapi.Identity{
					ObjectMeta: metav1.ObjectMeta{Name: "identity-no-user"},
					User:       kapi.ObjectReference{},
				},
				&authenticationapi.Identity{
					ObjectMeta: metav1.ObjectMeta{Name: "identity-matching-user"},
					User:       kapi.ObjectReference{Name: "bob"},
				},
				&authenticationapi.Identity{
					ObjectMeta: metav1.ObjectMeta{Name: "identity-different-uid"},
					User:       kapi.ObjectReference{Name: "bob", UID: "123"},
				},
				&authenticationapi.Identity{
					ObjectMeta: metav1.ObjectMeta{Name: "identity-different-user"},
					User:       kapi.ObjectReference{Name: "bob2"},
				},
			},
			expected: []interface{}{
				// Make sure identities are not messed with, only the user is removed
				clientgotesting.DeleteActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "delete", Resource: usersResource}, Name: "bob"},
			},
		},
		{
			name: "groups",
			user: "bob",
			objects: []runtime.Object{
				&authenticationapi.Group{
					ObjectMeta: metav1.ObjectMeta{Name: "group-no-users"},
					Users:      []string{},
				},
				&authenticationapi.Group{
					ObjectMeta: metav1.ObjectMeta{Name: "group-one-user"},
					Users:      []string{"bob"},
				},
				&authenticationapi.Group{
					ObjectMeta: metav1.ObjectMeta{Name: "group-multiple-users"},
					Users:      []string{"bob2", "bob", "steve"},
				},
				&authenticationapi.Group{
					ObjectMeta: metav1.ObjectMeta{Name: "group-mismatched-users"},
					Users:      []string{"bob2", "steve"},
				},
			},
			expected: []interface{}{
				clientgotesting.UpdateActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "update", Resource: groupsResource}, Object: &authenticationapi.Group{
					ObjectMeta: metav1.ObjectMeta{Name: "group-one-user"},
					Users:      []string{},
				}},
				clientgotesting.UpdateActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "update", Resource: groupsResource}, Object: &authenticationapi.Group{
					ObjectMeta: metav1.ObjectMeta{Name: "group-multiple-users"},
					Users:      []string{"bob2", "steve"},
				}},
				clientgotesting.DeleteActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "delete", Resource: usersResource}, Name: "bob"},
			},
		},
		{
			name: "oauth client authorizations",
			user: "bob",
			objects: []runtime.Object{
				&oauthapi.OAuthClientAuthorization{
					ObjectMeta: metav1.ObjectMeta{Name: "other-user"},
					UserName:   "alice",
					UserUID:    "123",
				},
				&oauthapi.OAuthClientAuthorization{
					ObjectMeta: metav1.ObjectMeta{Name: "bob-authorization-1"},
					UserName:   "bob",
					UserUID:    "234",
				},
				&oauthapi.OAuthClientAuthorization{
					ObjectMeta: metav1.ObjectMeta{Name: "bob-authorization-2"},
					UserName:   "bob",
					UserUID:    "345",
				},
			},
			expected: []interface{}{
				clientgotesting.DeleteActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "delete", Resource: oAuthClientAuthorizationsResource}, Name: "bob-authorization-1"},
				clientgotesting.DeleteActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "delete", Resource: oAuthClientAuthorizationsResource}, Name: "bob-authorization-2"},
				clientgotesting.DeleteActionImpl{ActionImpl: clientgotesting.ActionImpl{Verb: "delete", Resource: usersResource}, Name: "bob"},
			},
		},
	}

	for _, test := range tests {
		tc := testclient.NewSimpleFake(testclient.OriginObjects(test.objects)...)
		ktc := legacyclient.NewSimpleFake(test.sccs...)

		actual := []interface{}{}
		oreactor := func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			actual = append(actual, action)
			return false, nil, nil
		}
		kreactor := func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			actual = append(actual, action)
			return false, nil, nil
		}

		tc.PrependReactor("update", "*", oreactor)
		tc.PrependReactor("delete", "*", oreactor)
		ktc.Fake.PrependReactor("update", "*", kreactor)
		ktc.Fake.PrependReactor("delete", "*", kreactor)

		reaper := NewUserReaper(tc, tc, tc, tc, tc, ktc)
		err := reaper.Stop("", test.user, 0, nil)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
		}

		if !reflect.DeepEqual(test.expected, actual) {
			for i, x := range test.expected {
				t.Logf("Expected %d: %s", i, spew.Sprint(x))
			}
			for i, x := range actual {
				t.Logf("Actual %d:   %s", i, spew.Sprint(x))
			}
			t.Errorf("%s: unexpected actions", test.name)
		}
	}
}
