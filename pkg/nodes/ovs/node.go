package ovs

import (
	"fmt"
	"strings"

	"github.com/ai4networks/net4me/pkg/node"
	"github.com/ai4networks/net4me/pkg/port"
	"github.com/neaas/neslink"
	"github.com/vishvananda/netlink"
)

type Node struct {
	manager *Manager
	index   int
}

func (m *Manager) newNode(index int) *Node {
	return &Node{
		manager: m,
		index:   index,
	}
}

func (n *Node) ID() string {
	return fmt.Sprintf("%d", n.index)
}

func (n *Node) Manager() node.Manager {
	return n.manager
}

func (n *Node) Name() (string, error) {
	var link netlink.Link
	err := neslink.Do(
		n.manager.workingNetNs,
		neslink.NAGetLink(neslink.LPIndex(n.index), &link),
	)
	if err != nil {
		return "", fmt.Errorf("could not get link: %w", err)
	}
	return link.Attrs().Name, nil
}

func (n *Node) Device() string {
	return n.manager.Device()
}

func (n *Node) Start() error {
	if err := neslink.Do(
		n.manager.workingNetNs,
		neslink.LASetUp(neslink.LPIndex(n.index)),
	); err != nil {
		return fmt.Errorf("could not set the network bridge up with index %d: %w", n.index, err)
	}
	return nil
}

func (n *Node) Stop() error {
	if err := neslink.Do(
		n.manager.workingNetNs,
		neslink.LASetDown(neslink.LPIndex(n.index)),
	); err != nil {
		return fmt.Errorf("could not set the network bridge down with index %d: %w", n.index, err)
	}
	return nil
}

func (n *Node) Running() bool {
	var link netlink.Link
	err := neslink.Do(
		n.manager.workingNetNs,
		neslink.NAGetLink(neslink.LPIndex(n.index), &link),
	)
	if err != nil {
		return false
	}
	return strings.Contains(link.Attrs().Flags.String(), "up")
}

func (n *Node) Info() (map[string]interface{}, error) {
	var link netlink.Link
	err := neslink.Do(
		n.manager.workingNetNs,
		neslink.NAGetLink(neslink.LPIndex(n.index), &link),
	)
	if err != nil {
		return nil, fmt.Errorf("could not get link: %w", err)
	}
	return map[string]interface{}{
		"index": link.Attrs().Index,
		"name":  link.Attrs().Name,
		"hw":    link.Attrs().HardwareAddr,
		"mtu":   link.Attrs().MTU,
	}, nil
}

func (n *Node) NetNs() neslink.NsProvider {
	return n.manager.workingNetNs
}

func (n *Node) Ports() ([]port.Port, error) {
	bridgeName, err := n.Name()
	if err != nil {
		return nil, fmt.Errorf("could not get bridge name required for port fetch: %w", err)
	}
	portNames, err := n.manager.clientOvS.VSwitch.ListPorts(bridgeName)
	if err != nil {
		return nil, fmt.Errorf("could not list ports: %w", err)
	}
	ports := make([]port.Port, 0)
	for _, portName := range portNames {
		port, err := port.FromName(n.manager.workingNetNs, portName)
		if err != nil {
			return nil, fmt.Errorf("could not get port: %w", err)
		}
		ports = append(ports, port)
	}
	return ports, nil
}

func (n *Node) PortAdd(p port.Port) error {
	bridgeName, err := n.Name()
	if err != nil {
		return fmt.Errorf("could not get bridge name required for port add: %w", err)
	}
	if err := port.TakePort(n.NetNs(), p); err != nil {
		return fmt.Errorf("could not take port from pool: %w", err)
	}
	if err := neslink.Do(
		n.NetNs(),
		neslink.LAGeneric("add-to-ovs", func() error {
			return n.manager.clientOvS.VSwitch.AddPort(bridgeName, p.Attrs().Name)
		}),
	); err != nil {
		return fmt.Errorf("could not add port to bridge: %w", err)
	}
	return nil
}

func (n *Node) PortRemove(p port.Port) error {
	bridgeName, err := n.Name()
	if err != nil {
		return fmt.Errorf("could not get bridge name required for port remove: %w", err)
	}
	if err := neslink.Do(
		n.NetNs(),
		neslink.LAGeneric("del-from-ovs", func() error {
			return n.manager.clientOvS.VSwitch.DeletePort(bridgeName, p.Attrs().Name)
		}),
	); err != nil {
		return fmt.Errorf("could not remove port to bridge: %w", err)
	}
	if err := port.GivePort(n.NetNs(), p); err != nil {
		return fmt.Errorf("could not give port back to pool: %w", err)
	}
	return nil
}

func (n *Node) Stats() (map[string]interface{}, error) {
	var link netlink.Link
	err := neslink.Do(
		n.manager.workingNetNs,
		neslink.NAGetLink(neslink.LPIndex(n.index), &link),
	)
	if err != nil {
		return nil, fmt.Errorf("could not get link: %w", err)
	}
	return map[string]interface{}{
		"mainstat":      float64(link.Attrs().Statistics.TxBytes),
		"secondarystat": float64(link.Attrs().Statistics.RxBytes),

		"rx_bytes":   link.Attrs().Statistics.RxBytes,
		"rx_packets": link.Attrs().Statistics.RxPackets,
		"tx_bytes":   link.Attrs().Statistics.TxBytes,
		"tx_packets": link.Attrs().Statistics.TxPackets,
		"rx_dropped": link.Attrs().Statistics.RxDropped,
		"tx_dropped": link.Attrs().Statistics.TxDropped,
		"rx_errors":  link.Attrs().Statistics.RxErrors,
		"tx_errors":  link.Attrs().Statistics.TxErrors,
		"collisions": link.Attrs().Statistics.Collisions,
	}, nil
}
