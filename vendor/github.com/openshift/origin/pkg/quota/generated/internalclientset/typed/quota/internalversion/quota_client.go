package internalversion

import (
	"github.com/openshift/origin/pkg/quota/generated/internalclientset/scheme"
	rest "k8s.io/client-go/rest"
)

type QuotaInterface interface {
	RESTClient() rest.Interface
	AppliedClusterResourceQuotasGetter
	ClusterResourceQuotasGetter
}

// QuotaClient is used to interact with features provided by the quota.openshift.io group.
type QuotaClient struct {
	restClient rest.Interface
}

func (c *QuotaClient) AppliedClusterResourceQuotas(namespace string) AppliedClusterResourceQuotaInterface {
	return newAppliedClusterResourceQuotas(c, namespace)
}

func (c *QuotaClient) ClusterResourceQuotas() ClusterResourceQuotaInterface {
	return newClusterResourceQuotas(c)
}

// NewForConfig creates a new QuotaClient for the given config.
func NewForConfig(c *rest.Config) (*QuotaClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &QuotaClient{client}, nil
}

// NewForConfigOrDie creates a new QuotaClient for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *QuotaClient {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new QuotaClient for the given RESTClient.
func New(c rest.Interface) *QuotaClient {
	return &QuotaClient{c}
}

func setConfigDefaults(config *rest.Config) error {
	g, err := scheme.Registry.Group("quota.openshift.io")
	if err != nil {
		return err
	}

	config.APIPath = "/apis"
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	if config.GroupVersion == nil || config.GroupVersion.Group != g.GroupVersion.Group {
		gv := g.GroupVersion
		config.GroupVersion = &gv
	}
	config.NegotiatedSerializer = scheme.Codecs

	if config.QPS == 0 {
		config.QPS = 5
	}
	if config.Burst == 0 {
		config.Burst = 10
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *QuotaClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
