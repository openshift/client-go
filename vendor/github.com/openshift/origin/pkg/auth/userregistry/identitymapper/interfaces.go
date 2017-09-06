package identitymapper

import (
	userapi "github.com/openshift/origin/pkg/user/apis/user"
)

type UserToGroupMapper interface {
	GroupsFor(username string) ([]*userapi.Group, error)
}

type NoopGroupMapper struct{}

func (n NoopGroupMapper) GroupsFor(username string) ([]*userapi.Group, error) {
	return []*userapi.Group{}, nil
}
