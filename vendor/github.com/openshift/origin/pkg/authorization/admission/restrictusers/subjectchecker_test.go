package restrictusers

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	otestclient "github.com/openshift/origin/pkg/client/testclient"
	userapi "github.com/openshift/origin/pkg/user/apis/user"
)

func mustNewSubjectChecker(t *testing.T, spec *authorizationapi.RoleBindingRestrictionSpec) SubjectChecker {
	checker, err := NewSubjectChecker(spec)
	if err != nil {
		t.Errorf("unexpected error from NewChecker: %v, spec: %#v", err, spec)
	}

	return checker
}

func TestSubjectCheckers(t *testing.T) {
	var (
		userBobRef = kapi.ObjectReference{
			Kind: authorizationapi.UserKind,
			Name: "Bob",
		}
		userAliceRef = kapi.ObjectReference{
			Kind: authorizationapi.UserKind,
			Name: "Alice",
		}
		groupRef = kapi.ObjectReference{
			Kind: authorizationapi.GroupKind,
			Name: "group",
		}
		serviceaccountRef = kapi.ObjectReference{
			Kind:      authorizationapi.ServiceAccountKind,
			Namespace: "namespace",
			Name:      "serviceaccount",
		}
		systemuserRef = kapi.ObjectReference{
			Kind: authorizationapi.SystemUserKind,
			Name: "system user",
		}
		systemgroupRef = kapi.ObjectReference{
			Kind: authorizationapi.SystemGroupKind,
			Name: "system group",
		}
		group = userapi.Group{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "group",
				Labels: map[string]string{"baz": "quux"},
			},
			Users: []string{userBobRef.Name},
		}
		objects = []runtime.Object{
			&userapi.User{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "Alice",
					Labels: map[string]string{"foo": "bar"},
				},
			},
			&userapi.User{
				ObjectMeta: metav1.ObjectMeta{Name: "Bob"},
				Groups:     []string{"group"},
			},
			&group,
			&kapi.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "serviceaccount",
					Labels:    map[string]string{"xyzzy": "thud"},
				},
			},
		}
	)

	testCases := []struct {
		name        string
		checker     SubjectChecker
		subject     kapi.ObjectReference
		shouldAllow bool
	}{
		{
			name: "allow system user by literal name match",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					UserRestriction: &authorizationapi.UserRestriction{
						Users: []string{systemuserRef.Name},
					},
				}),
			subject:     systemuserRef,
			shouldAllow: true,
		},
		{
			name: "allow system group by literal name match",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					GroupRestriction: &authorizationapi.GroupRestriction{
						Groups: []string{systemgroupRef.Name},
					},
				}),
			subject:     systemgroupRef,
			shouldAllow: true,
		},
		{
			name: "allow regular user by literal name match",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					UserRestriction: &authorizationapi.UserRestriction{
						Users: []string{userAliceRef.Name},
					},
				}),
			subject:     userAliceRef,
			shouldAllow: true,
		},
		{
			name: "allow regular user by group membership",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					UserRestriction: &authorizationapi.UserRestriction{
						Groups: []string{groupRef.Name},
					},
				}),
			subject:     userBobRef,
			shouldAllow: true,
		},
		{
			name: "prohibit regular user when another user matches on group membership",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					UserRestriction: &authorizationapi.UserRestriction{
						Groups: []string{groupRef.Name},
					},
				}),
			subject:     userAliceRef,
			shouldAllow: false,
		},
		{
			name: "allow regular user by label selector match",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					UserRestriction: &authorizationapi.UserRestriction{
						Selectors: []metav1.LabelSelector{
							{MatchLabels: map[string]string{"foo": "bar"}},
						},
					},
				}),
			subject:     userAliceRef,
			shouldAllow: true,
		},
		{
			name: "prohibit regular user when another user matches on label selector",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					UserRestriction: &authorizationapi.UserRestriction{
						Selectors: []metav1.LabelSelector{
							{MatchLabels: map[string]string{"foo": "bar"}},
						},
					},
				}),
			subject:     userBobRef,
			shouldAllow: false,
		},
		{
			name: "allow regular group by literal name match",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					GroupRestriction: &authorizationapi.GroupRestriction{
						Groups: []string{groupRef.Name},
					},
				}),
			subject:     groupRef,
			shouldAllow: true,
		},
		{
			name: "allow regular group by label selector match",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					GroupRestriction: &authorizationapi.GroupRestriction{
						Selectors: []metav1.LabelSelector{
							{MatchLabels: map[string]string{"baz": "quux"}},
						},
					},
				}),
			subject:     groupRef,
			shouldAllow: true,
		},
		{
			name: "allow service account with explicit namespace by match on literal name and explicit namespace",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					ServiceAccountRestriction: &authorizationapi.ServiceAccountRestriction{
						ServiceAccounts: []authorizationapi.ServiceAccountReference{
							{
								Name:      serviceaccountRef.Name,
								Namespace: serviceaccountRef.Namespace,
							},
						},
					},
				}),
			subject:     serviceaccountRef,
			shouldAllow: true,
		},
		{
			name: "allow service account with explicit namespace by match on literal name and implicit namespace",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					ServiceAccountRestriction: &authorizationapi.ServiceAccountRestriction{
						ServiceAccounts: []authorizationapi.ServiceAccountReference{
							{Name: serviceaccountRef.Name},
						},
					},
				}),
			subject:     serviceaccountRef,
			shouldAllow: true,
		},
		{
			name: "prohibit service account with explicit namespace where literal name matches but explicit namespace does not",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					ServiceAccountRestriction: &authorizationapi.ServiceAccountRestriction{
						ServiceAccounts: []authorizationapi.ServiceAccountReference{
							{
								Namespace: serviceaccountRef.Namespace,
								Name:      serviceaccountRef.Name,
							},
						},
					},
				}),
			subject: kapi.ObjectReference{
				Kind:      authorizationapi.ServiceAccountKind,
				Namespace: "othernamespace",
				Name:      serviceaccountRef.Name,
			},
			shouldAllow: false,
		},
		{
			name: "prohibit service account with explicit namespace where literal name matches but implicit namespace does not",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					ServiceAccountRestriction: &authorizationapi.ServiceAccountRestriction{
						ServiceAccounts: []authorizationapi.ServiceAccountReference{
							{Name: serviceaccountRef.Name},
						},
					},
				}),
			subject: kapi.ObjectReference{
				Kind:      authorizationapi.ServiceAccountKind,
				Namespace: "othernamespace",
				Name:      serviceaccountRef.Name,
			},
			shouldAllow: false,
		},
		{
			name: "allow service account with implicit namespace by match on literal name and explicit namespace",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					ServiceAccountRestriction: &authorizationapi.ServiceAccountRestriction{
						ServiceAccounts: []authorizationapi.ServiceAccountReference{
							{
								Name:      serviceaccountRef.Name,
								Namespace: serviceaccountRef.Namespace,
							},
						},
					},
				}),
			subject: kapi.ObjectReference{
				Kind: authorizationapi.ServiceAccountKind,
				Name: serviceaccountRef.Name,
			},
			shouldAllow: true,
		},
		{
			name: "allow service account with implicit namespace by match on literal name and implicit namespace",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					ServiceAccountRestriction: &authorizationapi.ServiceAccountRestriction{
						ServiceAccounts: []authorizationapi.ServiceAccountReference{
							{Name: serviceaccountRef.Name},
						},
					},
				}),
			subject: kapi.ObjectReference{
				Kind: authorizationapi.ServiceAccountKind,
				Name: serviceaccountRef.Name,
			},
			shouldAllow: true,
		},
		{
			name: "prohibit service account with implicit namespace where literal name matches but explicit namespace does not",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					ServiceAccountRestriction: &authorizationapi.ServiceAccountRestriction{
						ServiceAccounts: []authorizationapi.ServiceAccountReference{
							{
								Namespace: "othernamespace",
								Name:      serviceaccountRef.Name,
							},
						},
					},
				}),
			subject: kapi.ObjectReference{
				Kind: authorizationapi.ServiceAccountKind,
				Name: serviceaccountRef.Name,
			},
			shouldAllow: false,
		},
		{
			name: "prohibit service account with explicit namespace where explicit namespace matches but literal name does not",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					ServiceAccountRestriction: &authorizationapi.ServiceAccountRestriction{
						ServiceAccounts: []authorizationapi.ServiceAccountReference{
							{
								Namespace: serviceaccountRef.Namespace,
								Name:      "othername",
							},
						},
					},
				}),
			subject:     serviceaccountRef,
			shouldAllow: false,
		},
		{
			name: "allow service account by match on namespace",
			checker: mustNewSubjectChecker(t,
				&authorizationapi.RoleBindingRestrictionSpec{
					ServiceAccountRestriction: &authorizationapi.ServiceAccountRestriction{
						Namespaces: []string{serviceaccountRef.Namespace},
					},
				}),
			subject:     serviceaccountRef,
			shouldAllow: true,
		},
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	kclient := fake.NewSimpleClientset(otestclient.UpstreamObjects(objects)...)
	oclient := otestclient.NewSimpleFake(otestclient.OriginObjects(objects)...)
	groupCache := fakeGroupCache{groups: []userapi.Group{group}}
	// This is a terrible, horrible, no-good, very bad hack to avoid a race
	// condition between the test "allow regular user by group membership"
	// and the group cache's initialisation.
	for {
		if groups, _ := groupCache.GroupsFor(group.Users[0]); len(groups) == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	ctx, err := NewRoleBindingRestrictionContext("namespace",
		kclient, oclient, groupCache)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	for _, tc := range testCases {
		allowed, err := tc.checker.Allowed(tc.subject, ctx)
		if err != nil {
			t.Errorf("test case %v: unexpected error: %v", tc.name, err)
		}
		if allowed && !tc.shouldAllow {
			t.Errorf("test case %v: subject allowed but should be prohibited", tc.name)
		}
		if !allowed && tc.shouldAllow {
			t.Errorf("test case %v: subject prohibited but should be allowed", tc.name)
		}
	}
}
