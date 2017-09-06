package ovs

import (
	"fmt"
	"sort"
)

// ovsFake implements a fake ovs.Interface for testing purposes
//
// Note that the code here is *not* expected to be 100% equivalent to ovsExec, as
// that would require porting over huge amounts of ovs-ofctl source code. It needs
// to support enough features to make the SDN unit tests pass, and should do enough
// error checking to catch bugs that have tripped us up in the past (eg,
// specifying "nw_dst" without "ip").
type ovsFake struct {
	bridge string

	ports map[string]int
	flows ovsFlows
}

// NewFake returns a new ovs.Interface
func NewFake(bridge string) Interface {
	return &ovsFake{bridge: bridge}
}

func (fake *ovsFake) AddBridge(properties ...string) error {
	fake.ports = make(map[string]int)
	fake.flows = make([]OvsFlow, 0)
	return nil
}

func (fake *ovsFake) DeleteBridge() error {
	fake.ports = nil
	fake.flows = nil
	return nil
}

func (fake *ovsFake) ensureExists() error {
	if fake.ports == nil {
		return fmt.Errorf("no bridge named %s", fake.bridge)
	}
	return nil
}

func (fake *ovsFake) GetOFPort(port string) (int, error) {
	if err := fake.ensureExists(); err != nil {
		return -1, err
	}

	if ofport, exists := fake.ports[port]; exists {
		return ofport, nil
	} else {
		return -1, fmt.Errorf("no row %q in table Interface", port)
	}
}

func (fake *ovsFake) AddPort(port string, ofportRequest int, properties ...string) (int, error) {
	if err := fake.ensureExists(); err != nil {
		return -1, err
	}

	ofport, exists := fake.ports[port]
	if exists {
		if ofport != ofportRequest && ofportRequest != -1 {
			return -1, fmt.Errorf("allocated ofport (%d) did not match request (%d)", ofport, ofportRequest)
		}
	} else {
		if ofportRequest == -1 {
			ofport := 1
			for _, existingPort := range fake.ports {
				if existingPort >= ofport {
					ofport = existingPort + 1
				}
			}
		} else {
			if ofportRequest < 1 || ofportRequest > 65535 {
				return -1, fmt.Errorf("requested ofport (%d) out of range", ofportRequest)
			}
			ofport = ofportRequest
		}
		fake.ports[port] = ofport
	}
	return ofport, nil
}

func (fake *ovsFake) DeletePort(port string) error {
	if err := fake.ensureExists(); err != nil {
		return err
	}

	delete(fake.ports, port)
	return nil
}

func (fake *ovsFake) SetFrags(mode string) error {
	return nil
}

func (ovsif *ovsFake) Create(table string, values ...string) (string, error) {
	return "fake-UUID", nil
}

func (fake *ovsFake) Destroy(table, record string) error {
	return nil
}

func (fake *ovsFake) Get(table, record, column string) (string, error) {
	return "", nil
}

func (fake *ovsFake) Set(table, record string, values ...string) error {
	return nil
}

func (fake *ovsFake) Clear(table, record string, columns ...string) error {
	return nil
}

type ovsFakeTx struct {
	fake *ovsFake
	err  error
}

func (fake *ovsFake) NewTransaction() Transaction {
	return &ovsFakeTx{fake: fake, err: fake.ensureExists()}
}

// sort.Interface support
type ovsFlows []OvsFlow

func (f ovsFlows) Len() int      { return len(f) }
func (f ovsFlows) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f ovsFlows) Less(i, j int) bool {
	if f[i].Table != f[j].Table {
		return f[i].Table < f[j].Table
	}
	if f[i].Priority != f[j].Priority {
		return f[i].Priority > f[j].Priority
	}
	return f[i].Created.Before(f[j].Created)
}

func fixFlowFields(flow *OvsFlow) {
	// Fix up field names to match what dump-flows prints.  Some fields
	// have aliases or deprecated names that can be used for add/del flows,
	// but dump always reports the canonical name
	if _, isArp := flow.FindField("arp"); isArp {
		for i := range flow.Fields {
			if flow.Fields[i].Name == "nw_src" {
				flow.Fields[i].Name = "arp_spa"
			} else if flow.Fields[i].Name == "nw_dst" {
				flow.Fields[i].Name = "arp_tpa"
			}
		}
	}
}

func (tx *ovsFakeTx) AddFlow(flow string, args ...interface{}) {
	if tx.err != nil {
		return
	}
	parsed, err := ParseFlow(ParseForAdd, flow, args...)
	if err != nil {
		tx.err = err
		return
	}
	fixFlowFields(parsed)

	// If there is already an exact match for this flow, then the new flow replaces it.
	for i := range tx.fake.flows {
		if FlowMatches(&tx.fake.flows[i], parsed, true) {
			tx.fake.flows[i] = *parsed
			return
		}
	}

	tx.fake.flows = append(tx.fake.flows, *parsed)
	sort.Sort(ovsFlows(tx.fake.flows))
}

func (tx *ovsFakeTx) DeleteFlows(flow string, args ...interface{}) {
	if tx.err != nil {
		return
	}
	parsed, err := ParseFlow(ParseForDelete, flow, args...)
	if err != nil {
		tx.err = err
		return
	}
	fixFlowFields(parsed)

	newFlows := make([]OvsFlow, 0, len(tx.fake.flows))
	for _, flow := range tx.fake.flows {
		if !FlowMatches(&flow, parsed, false) {
			newFlows = append(newFlows, flow)
		}
	}
	tx.fake.flows = newFlows
}

func (tx *ovsFakeTx) EndTransaction() error {
	err := tx.err
	tx.err = nil
	return err
}

func (fake *ovsFake) DumpFlows() ([]string, error) {
	if err := fake.ensureExists(); err != nil {
		return nil, err
	}

	flows := make([]string, 0, len(fake.flows))
	for _, flow := range fake.flows {
		str := fmt.Sprintf(" cookie=%s, table=%d", flow.Cookie, flow.Table)
		if flow.Priority != defaultPriority {
			str += fmt.Sprintf(", priority=%d", flow.Priority)
		}
		for _, field := range flow.Fields {
			if field.Value == "" {
				str += fmt.Sprintf(", %s", field.Name)
			} else {
				str += fmt.Sprintf(", %s=%s", field.Name, field.Value)
			}
		}
		actionStr := ""
		for _, action := range flow.Actions {
			if len(actionStr) > 0 {
				actionStr += ","
			}
			actionStr += action.Name
			if action.Value != "" {
				if action.Value[0] != '(' {
					actionStr += ":" + action.Value
				} else {
					actionStr += action.Value
				}
			}
		}
		if len(actionStr) > 0 {
			str += fmt.Sprintf(", actions=%s", actionStr)
		}
		flows = append(flows, str)
	}

	return flows, nil
}
