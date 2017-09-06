package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ClusterNetworkDefault = "default"
)

// +genclient
// +genclient:nonNamespaced

// ClusterNetwork describes the cluster network. There is normally only one object of this type,
// named "default", which is created by the SDN network plugin based on the master configuration
// when the cluster is brought up for the first time.
type ClusterNetwork struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Network is a CIDR string specifying the global overlay network's L3 space
	Network string `json:"network" protobuf:"bytes,2,opt,name=network"`
	// HostSubnetLength is the number of bits of network to allocate to each node. eg, 8 would mean that each node would have a /24 slice of the overlay network for its pods
	HostSubnetLength uint32 `json:"hostsubnetlength" protobuf:"varint,3,opt,name=hostsubnetlength"`
	// ServiceNetwork is the CIDR range that Service IP addresses are allocated from
	ServiceNetwork string `json:"serviceNetwork" protobuf:"bytes,4,opt,name=serviceNetwork"`
	// PluginName is the name of the network plugin being used
	PluginName string `json:"pluginName,omitempty" protobuf:"bytes,5,opt,name=pluginName"`
}

// ClusterNetworkList is a collection of ClusterNetworks
type ClusterNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is the list of cluster networks
	Items []ClusterNetwork `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced

// HostSubnet describes the container subnet network on a node. The HostSubnet object must have the
// same name as the Node object it corresponds to.
type HostSubnet struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Host is the name of the node. (This is the same as the object's name, but both fields must be set.)
	Host string `json:"host" protobuf:"bytes,2,opt,name=host"`
	// HostIP is the IP address to be used as a VTEP by other nodes in the overlay network
	HostIP string `json:"hostIP" protobuf:"bytes,3,opt,name=hostIP"`
	// Subnet is the CIDR range of the overlay network assigned to the node for its pods
	Subnet string `json:"subnet" protobuf:"bytes,4,opt,name=subnet"`
}

// HostSubnetList is a collection of HostSubnets
type HostSubnetList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is the list of host subnets
	Items []HostSubnet `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced

// NetNamespace describes a single isolated network. When using the redhat/openshift-ovs-multitenant
// plugin, every Namespace will have a corresponding NetNamespace object with the same name.
// (When using redhat/openshift-ovs-subnet, NetNamespaces are not used.)
type NetNamespace struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// NetName is the name of the network namespace. (This is the same as the object's name, but both fields must be set.)
	NetName string `json:"netname" protobuf:"bytes,2,opt,name=netname"`
	// NetID is the network identifier of the network namespace assigned to each overlay network packet. This can be manipulated with the "oc adm pod-network" commands.
	NetID uint32 `json:"netid" protobuf:"varint,3,opt,name=netid"`
}

// NetNamespaceList is a collection of NetNamespaces
type NetNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is the list of net namespaces
	Items []NetNamespace `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// EgressNetworkPolicyRuleType indicates whether an EgressNetworkPolicyRule allows or denies traffic
type EgressNetworkPolicyRuleType string

const (
	EgressNetworkPolicyRuleAllow EgressNetworkPolicyRuleType = "Allow"
	EgressNetworkPolicyRuleDeny  EgressNetworkPolicyRuleType = "Deny"
)

// EgressNetworkPolicyPeer specifies a target to apply egress network policy to
type EgressNetworkPolicyPeer struct {
	// cidrSelector is the CIDR range to allow/deny traffic to. If this is set, dnsName must be unset
	CIDRSelector string `json:"cidrSelector,omitempty" protobuf:"bytes,1,rep,name=cidrSelector"`
	// dnsName is the domain name to allow/deny traffic to. If this is set, cidrSelector must be unset
	DNSName string `json:"dnsName,omitempty" protobuf:"bytes,2,rep,name=dnsName"`
}

// EgressNetworkPolicyRule contains a single egress network policy rule
type EgressNetworkPolicyRule struct {
	// type marks this as an "Allow" or "Deny" rule
	Type EgressNetworkPolicyRuleType `json:"type" protobuf:"bytes,1,rep,name=type"`
	// to is the target that traffic is allowed/denied to
	To EgressNetworkPolicyPeer `json:"to" protobuf:"bytes,2,rep,name=to"`
}

// EgressNetworkPolicySpec provides a list of policies on outgoing network traffic
type EgressNetworkPolicySpec struct {
	// egress contains the list of egress policy rules
	Egress []EgressNetworkPolicyRule `json:"egress" protobuf:"bytes,1,rep,name=egress"`
}

// +genclient

// EgressNetworkPolicy describes the current egress network policy for a Namespace. When using
// the 'redhat/openshift-ovs-multitenant' network plugin, traffic from a pod to an IP address
// outside the cluster will be checked against each EgressNetworkPolicyRule in the pod's
// namespace's EgressNetworkPolicy, in order. If no rule matches (or no EgressNetworkPolicy
// is present) then the traffic will be allowed by default.
type EgressNetworkPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// metadata for EgressNetworkPolicy
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec is the specification of the current egress network policy
	Spec EgressNetworkPolicySpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
}

// EgressNetworkPolicyList is a collection of EgressNetworkPolicy
type EgressNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	// metadata for EgressNetworkPolicyList
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// items is the list of policies
	Items []EgressNetworkPolicy `json:"items" protobuf:"bytes,2,rep,name=items"`
}
