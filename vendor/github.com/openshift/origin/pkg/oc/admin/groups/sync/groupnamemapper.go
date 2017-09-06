package syncgroups

import (
	"errors"
	"fmt"

	kutilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/openshift/origin/pkg/auth/ldaputil"
	"github.com/openshift/origin/pkg/oc/admin/groups/sync/interfaces"
)

// NewUserDefinedGroupNameMapper returns a new UserDefinedLDAPGroupNameMapper which maps a ldapGroupUID
// representing an LDAP group to the OpenShift Group name for the resource
func NewUserDefinedGroupNameMapper(mapping map[string]string) interfaces.LDAPGroupNameMapper {
	return &UserDefinedLDAPGroupNameMapper{
		nameMapping: mapping,
	}
}

// UserDefinedLDAPGroupNameMapper maps a ldapGroupUID representing an LDAP group to the OpenShift Group
// name for the resource by using a pre-defined mapping of ldapGroupUID to name (e.g. from a file)
type UserDefinedLDAPGroupNameMapper struct {
	nameMapping map[string]string
}

func (m *UserDefinedLDAPGroupNameMapper) GroupNameFor(ldapGroupUID string) (string, error) {
	openShiftGroupName, exists := m.nameMapping[ldapGroupUID]
	if !exists {
		return "", fmt.Errorf("no OpenShift Group name defined for LDAP group UID: %s", ldapGroupUID)
	}
	return openShiftGroupName, nil
}

// NewEntryAttributeGroupNameMapper returns a new EntryAttributeLDAPGroupNameMapper
func NewEntryAttributeGroupNameMapper(nameAttribute []string, groupGetter interfaces.LDAPGroupGetter) interfaces.LDAPGroupNameMapper {
	return &EntryAttributeLDAPGroupNameMapper{
		nameAttribute: nameAttribute,
		groupGetter:   groupGetter,
	}
}

// EntryAttributeLDAPGroupNameMapper references the name attribute mapping to determine which attribute
// of a first-class LDAP group entry should be used as the OpenShift Group name for the resource
type EntryAttributeLDAPGroupNameMapper struct {
	nameAttribute []string
	groupGetter   interfaces.LDAPGroupGetter
}

func (m *EntryAttributeLDAPGroupNameMapper) GroupNameFor(ldapGroupUID string) (string, error) {
	group, err := m.groupGetter.GroupEntryFor(ldapGroupUID)
	if err != nil {
		return "", err
	}
	openShiftGroupName := ldaputil.GetAttributeValue(group, m.nameAttribute)
	if len(openShiftGroupName) == 0 {
		return "",
			fmt.Errorf("the group entry (%v: %v) does not map to an OpenShift Group name with the given name attribute (%v)", group, group.Attributes, m.nameAttribute)
	}
	return openShiftGroupName, nil
}

// DNLDAPGroupNameMapper passes through the ldapGroupUID as the OpenShift group name
type DNLDAPGroupNameMapper struct{}

func (m *DNLDAPGroupNameMapper) GroupNameFor(ldapGroupUID string) (string, error) {
	return ldapGroupUID, nil
}

type UnionGroupNameMapper struct {
	GroupNameMappers []interfaces.LDAPGroupNameMapper
}

func (m *UnionGroupNameMapper) GroupNameFor(ldapGroupUID string) (string, error) {
	if len(m.GroupNameMappers) == 0 {
		return "", errors.New("no group name mappers defined")
	}

	errs := []error{}
	for _, currMapper := range m.GroupNameMappers {
		ret, err := currMapper.GroupNameFor(ldapGroupUID)
		if err == nil {
			return ret, nil
		}

		errs = append(errs, err)
	}

	return "", kutilerrors.NewAggregate(errs)
}
