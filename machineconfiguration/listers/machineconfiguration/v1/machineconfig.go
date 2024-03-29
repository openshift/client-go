// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/openshift/api/machineconfiguration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MachineConfigLister helps list MachineConfigs.
// All objects returned here must be treated as read-only.
type MachineConfigLister interface {
	// List lists all MachineConfigs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.MachineConfig, err error)
	// Get retrieves the MachineConfig from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.MachineConfig, error)
	MachineConfigListerExpansion
}

// machineConfigLister implements the MachineConfigLister interface.
type machineConfigLister struct {
	indexer cache.Indexer
}

// NewMachineConfigLister returns a new MachineConfigLister.
func NewMachineConfigLister(indexer cache.Indexer) MachineConfigLister {
	return &machineConfigLister{indexer: indexer}
}

// List lists all MachineConfigs in the indexer.
func (s *machineConfigLister) List(selector labels.Selector) (ret []*v1.MachineConfig, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.MachineConfig))
	})
	return ret, err
}

// Get retrieves the MachineConfig from the index for a given name.
func (s *machineConfigLister) Get(name string) (*v1.MachineConfig, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("machineconfig"), name)
	}
	return obj.(*v1.MachineConfig), nil
}
