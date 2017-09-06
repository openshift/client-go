package admission

import (
	"net"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
	kapi "k8s.io/kubernetes/pkg/api"
)

// TestAdmission verifies various scenarios involving pod/project/global node label selectors
func TestAdmission(t *testing.T) {
	svc := &kapi.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}
	var oldSvc *kapi.Service

	_, ipv4, err := net.ParseCIDR("172.0.0.0/16")
	if err != nil {
		t.Fatal(err)
	}
	_, ipv4subset, err := net.ParseCIDR("172.0.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	_, ipv4offset, err := net.ParseCIDR("172.200.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	_, none, err := net.ParseCIDR("0.0.0.0/32")
	if err != nil {
		t.Fatal(err)
	}
	_, all, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		testName        string
		rejects, admits []*net.IPNet
		op              admission.Operation
		externalIPs     []string
		admit           bool
		errFn           func(err error) bool
		loadBalancer    bool
		ingressIP       string
	}{
		{
			admit:    true,
			op:       admission.Create,
			testName: "No external IPs on create",
		},
		{
			admit:    true,
			op:       admission.Update,
			testName: "No external IPs on update",
		},
		{
			admit:       false,
			externalIPs: []string{"1.2.3.4"},
			op:          admission.Create,
			testName:    "No external IPs allowed on create",
			errFn:       func(err error) bool { return strings.Contains(err.Error(), "externalIPs have been disabled") },
		},
		{
			admit:       false,
			externalIPs: []string{"1.2.3.4"},
			op:          admission.Update,
			testName:    "No external IPs allowed on update",
			errFn:       func(err error) bool { return strings.Contains(err.Error(), "externalIPs have been disabled") },
		},
		{
			admit:       false,
			admits:      []*net.IPNet{ipv4},
			externalIPs: []string{"1.2.3.4"},
			op:          admission.Create,
			testName:    "IP out of range on create",
			errFn: func(err error) bool {
				return strings.Contains(err.Error(), "externalIP is not allowed") &&
					strings.Contains(err.Error(), "spec.externalIPs[0]")
			},
		},
		{
			admit:       false,
			admits:      []*net.IPNet{ipv4},
			externalIPs: []string{"1.2.3.4"},
			op:          admission.Update,
			testName:    "IP out of range on update",
			errFn: func(err error) bool {
				return strings.Contains(err.Error(), "externalIP is not allowed") &&
					strings.Contains(err.Error(), "spec.externalIPs[0]")
			},
		},
		{
			admit:       false,
			admits:      []*net.IPNet{ipv4},
			rejects:     []*net.IPNet{ipv4subset},
			externalIPs: []string{"172.0.1.1"},
			op:          admission.Update,
			testName:    "IP out of range due to blacklist",
			errFn: func(err error) bool {
				return strings.Contains(err.Error(), "externalIP is not allowed") &&
					strings.Contains(err.Error(), "spec.externalIPs[0]")
			},
		},
		{
			admit:       false,
			admits:      []*net.IPNet{ipv4},
			rejects:     []*net.IPNet{ipv4offset},
			externalIPs: []string{"172.199.1.1"},
			op:          admission.Update,
			testName:    "IP not in reject or admit",
			errFn: func(err error) bool {
				return strings.Contains(err.Error(), "externalIP is not allowed") &&
					strings.Contains(err.Error(), "spec.externalIPs[0]")
			},
		},
		{
			admit:       true,
			admits:      []*net.IPNet{ipv4},
			externalIPs: []string{"172.0.0.1"},
			op:          admission.Create,
			testName:    "IP in range on create",
		},
		{
			admit:       true,
			admits:      []*net.IPNet{ipv4},
			externalIPs: []string{"172.0.0.1"},
			op:          admission.Update,
			testName:    "IP in range on update",
		},
		// other checks
		{
			admit:       false,
			admits:      []*net.IPNet{ipv4},
			externalIPs: []string{"abcd"},
			op:          admission.Create,
			testName:    "IP unparseable on create",
			errFn: func(err error) bool {
				return strings.Contains(err.Error(), "externalIPs must be a valid address") &&
					strings.Contains(err.Error(), "spec.externalIPs[0]")
			},
		},
		{
			admit:       false,
			admits:      []*net.IPNet{none},
			externalIPs: []string{"1.2.3.4"},
			op:          admission.Create,
			testName:    "IP range is empty",
		},
		{
			admit:       false,
			rejects:     []*net.IPNet{all},
			admits:      []*net.IPNet{all},
			externalIPs: []string{"1.2.3.4"},
			op:          admission.Create,
			testName:    "rejections can cover the entire range",
		},
		// Ingress IP checks
		{
			admit:        true,
			externalIPs:  []string{"1.2.3.4"},
			op:           admission.Update,
			testName:     "Ingress ip allowed when external ips are disabled",
			loadBalancer: true,
			ingressIP:    "1.2.3.4",
		},
		{
			admit:        true,
			admits:       []*net.IPNet{ipv4},
			externalIPs:  []string{"1.2.3.4", "172.0.0.1"},
			op:           admission.Update,
			testName:     "Ingress ip allowed when external ips are enabled",
			loadBalancer: true,
			ingressIP:    "1.2.3.4",
		},
		{
			admit:        false,
			admits:       []*net.IPNet{ipv4},
			externalIPs:  []string{"1.2.3.4", "172.0.0.1"},
			op:           admission.Update,
			testName:     "Ingress ip not allowed for non-lb service",
			loadBalancer: false,
			ingressIP:    "1.2.3.4",
		},
	}
	for _, test := range tests {
		svc.Spec.ExternalIPs = test.externalIPs
		allowIngressIP := len(test.ingressIP) > 0 || test.loadBalancer
		handler := NewExternalIPRanger(test.rejects, test.admits, allowIngressIP)

		if test.loadBalancer {
			svc.Spec.Type = kapi.ServiceTypeLoadBalancer
		} else {
			svc.Spec.Type = kapi.ServiceTypeClusterIP
		}

		if len(test.ingressIP) > 0 {
			// Provide an ingress ip via the previous object state
			oldSvc = &kapi.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Status: kapi.ServiceStatus{
					LoadBalancer: kapi.LoadBalancerStatus{
						Ingress: []kapi.LoadBalancerIngress{
							{
								IP: test.ingressIP,
							},
						},
					},
				},
			}

		} else {
			oldSvc = nil
		}

		err := handler.Admit(admission.NewAttributesRecord(svc, oldSvc, kapi.Kind("Service").WithVersion("version"), "namespace", svc.ObjectMeta.Name, kapi.Resource("services").WithVersion("version"), "", test.op, nil))
		if test.admit && err != nil {
			t.Errorf("%s: expected no error but got: %s", test.testName, err)
		} else if !test.admit && err == nil {
			t.Errorf("%s: expected an error", test.testName)
		}
		if test.errFn != nil && !test.errFn(err) {
			t.Errorf("%s: unexpected error: %v", test.testName, err)
		}
	}
}

func TestHandles(t *testing.T) {
	for op, shouldHandle := range map[admission.Operation]bool{
		admission.Create:  true,
		admission.Update:  true,
		admission.Connect: false,
		admission.Delete:  false,
	} {
		ranger := NewExternalIPRanger(nil, nil, false)
		if e, a := shouldHandle, ranger.Handles(op); e != a {
			t.Errorf("%v: shouldHandle=%t, handles=%t", op, e, a)
		}
	}
}
