// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/openshift/api/security/v1"
	securityv1 "github.com/openshift/client-go/security/applyconfigurations/security/v1"
	typedsecurityv1 "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	gentype "k8s.io/client-go/gentype"
)

// fakeSecurityContextConstraints implements SecurityContextConstraintsInterface
type fakeSecurityContextConstraints struct {
	*gentype.FakeClientWithListAndApply[*v1.SecurityContextConstraints, *v1.SecurityContextConstraintsList, *securityv1.SecurityContextConstraintsApplyConfiguration]
	Fake *FakeSecurityV1
}

func newFakeSecurityContextConstraints(fake *FakeSecurityV1) typedsecurityv1.SecurityContextConstraintsInterface {
	return &fakeSecurityContextConstraints{
		gentype.NewFakeClientWithListAndApply[*v1.SecurityContextConstraints, *v1.SecurityContextConstraintsList, *securityv1.SecurityContextConstraintsApplyConfiguration](
			fake.Fake,
			"",
			v1.SchemeGroupVersion.WithResource("securitycontextconstraints"),
			v1.SchemeGroupVersion.WithKind("SecurityContextConstraints"),
			func() *v1.SecurityContextConstraints { return &v1.SecurityContextConstraints{} },
			func() *v1.SecurityContextConstraintsList { return &v1.SecurityContextConstraintsList{} },
			func(dst, src *v1.SecurityContextConstraintsList) { dst.ListMeta = src.ListMeta },
			func(list *v1.SecurityContextConstraintsList) []*v1.SecurityContextConstraints {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1.SecurityContextConstraintsList, items []*v1.SecurityContextConstraints) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
