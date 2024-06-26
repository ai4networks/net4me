package ovs

import (
	"fmt"
	"maps"
	"os"
	"sync"

	"github.com/ai4networks/net4me/pkg/node"
	"github.com/mitchellh/mapstructure"
	"github.com/neaas/go-openvswitch/ovs"
	"github.com/neaas/neslink"
	"github.com/vishvananda/netlink"
)

type Manager struct {
	lock         *sync.RWMutex
	clientOvS    *ovs.Client
	workingNetNs neslink.NsProvider
}

type ManagerConfig struct {
	Sudo bool `mapstructure:"sudo"`
}

type AddConfig struct {
	ControllerIP string `mapstructure:"controller_ip"`
}

func (m *Manager) Device() string {
	return "ovs"
}

func (m *Manager) Setup(config map[string]interface{}) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	defaultConfig := map[string]any{
		"sudo": false, // TODO: this should be false when out of initial dev phase
	}
	maps.Copy(defaultConfig, config)
	c := ManagerConfig{}
	if err := mapstructure.Decode(defaultConfig, &c); err != nil {
		return fmt.Errorf("failed to decode ovs manager config: %w", err)
	}
	if c.Sudo {
		m.clientOvS = ovs.New(ovs.Sudo())
	} else {
		m.clientOvS = ovs.New()
	}
	m.workingNetNs = neslink.NPProcess(os.Getpid())
	return nil
}

func (m *Manager) Info() (map[string]any, error) {
	return map[string]any{}, nil
}

func (m *Manager) Icon() string {
	return "gf-glue"
}

func (m *Manager) Color() string {
	return "LightSalmon"
}

func (m *Manager) Nodes(nodeFilters ...node.NodeFilter) ([]node.Node, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	bridges, err := m.clientOvS.VSwitch.ListBridges()
	if err != nil {
		return nil, fmt.Errorf("could not list bridges: %w", err)
	}
	nodes := make([]node.Node, 0)
	for _, bridge := range bridges {
		var link netlink.Link
		if err := neslink.Do(
			m.workingNetNs,
			neslink.NAGetLink(neslink.LPName(bridge), &link),
		); err != nil {
			return nil, fmt.Errorf("could not get link of new bridge: %w", err)
		}
		nodes = append(nodes, m.newNode(link.Attrs().Index))
	}
	return nodes, nil
}

func (m *Manager) Add(name string, labels map[string]string, config map[string]any) (node.Node, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var c AddConfig
	if err := mapstructure.Decode(config, &c); err != nil {
		return nil, fmt.Errorf("failed to decode ovs add config: %w", err)
	}
	// TODO: add labels somehow? maybe a file of other db solution?
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if err := neslink.Do(
		m.workingNetNs,
		neslink.LAGeneric("add-ovs-bridge", func() error {
			return m.clientOvS.VSwitch.AddBridge(name)
		}),
	); err != nil {
		return nil, fmt.Errorf("could not create network bridge %s: %w", name, err)
	}
	if err := neslink.Do(
		m.workingNetNs,
		neslink.LASetDown(neslink.LPName(name)),
	); err != nil {
		return nil, fmt.Errorf("could not set the network bridge down %s: %w", name, err)
	}
	var link netlink.Link
	if err := neslink.Do(
		m.workingNetNs,
		neslink.NAGetLink(neslink.LPName(name), &link),
	); err != nil {
		return nil, fmt.Errorf("bridge created, but could not be found %s: %w", name, err)
	}
	if c.ControllerIP != "" {
		if err := neslink.Do(
			m.workingNetNs,
			neslink.LAGeneric("add-ovs-controller", func() error {
				return m.clientOvS.VSwitch.SetController(name, c.ControllerIP)
			}),
		); err != nil {
			return nil, fmt.Errorf("could not add controller to network bridge %s: %w", name, err)
		}
	}
	return m.newNode(link.Attrs().Index), nil
}

func (m *Manager) Remove(n node.Node) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	name, err := n.Name()
	if err != nil {
		return fmt.Errorf("could not remove network bridge %s since name could not be determined: %w", n.ID(), err)
	}
	if err := neslink.Do(
		m.workingNetNs,
		neslink.LAGeneric("del-ovs-bridge", func() error {
			return m.clientOvS.VSwitch.DeleteBridge(name)
		}),
	); err != nil {
		return fmt.Errorf("could not remove network bridge %s: %w", name, err)
	}
	return nil
}

func init() {
	node.RegisterManager(&Manager{
		lock: &sync.RWMutex{},
	})
}
