// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"

	v1 "github.com/openshift/api/monitoring/v1"
	monitoringv1 "github.com/openshift/client-go/monitoring/applyconfigurations/monitoring/v1"
	scheme "github.com/openshift/client-go/monitoring/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// AlertingRulesGetter has a method to return a AlertingRuleInterface.
// A group's client should implement this interface.
type AlertingRulesGetter interface {
	AlertingRules(namespace string) AlertingRuleInterface
}

// AlertingRuleInterface has methods to work with AlertingRule resources.
type AlertingRuleInterface interface {
	Create(ctx context.Context, alertingRule *v1.AlertingRule, opts metav1.CreateOptions) (*v1.AlertingRule, error)
	Update(ctx context.Context, alertingRule *v1.AlertingRule, opts metav1.UpdateOptions) (*v1.AlertingRule, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, alertingRule *v1.AlertingRule, opts metav1.UpdateOptions) (*v1.AlertingRule, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.AlertingRule, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.AlertingRuleList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.AlertingRule, err error)
	Apply(ctx context.Context, alertingRule *monitoringv1.AlertingRuleApplyConfiguration, opts metav1.ApplyOptions) (result *v1.AlertingRule, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, alertingRule *monitoringv1.AlertingRuleApplyConfiguration, opts metav1.ApplyOptions) (result *v1.AlertingRule, err error)
	AlertingRuleExpansion
}

// alertingRules implements AlertingRuleInterface
type alertingRules struct {
	*gentype.ClientWithListAndApply[*v1.AlertingRule, *v1.AlertingRuleList, *monitoringv1.AlertingRuleApplyConfiguration]
}

// newAlertingRules returns a AlertingRules
func newAlertingRules(c *MonitoringV1Client, namespace string) *alertingRules {
	return &alertingRules{
		gentype.NewClientWithListAndApply[*v1.AlertingRule, *v1.AlertingRuleList, *monitoringv1.AlertingRuleApplyConfiguration](
			"alertingrules",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1.AlertingRule { return &v1.AlertingRule{} },
			func() *v1.AlertingRuleList { return &v1.AlertingRuleList{} }),
	}
}
