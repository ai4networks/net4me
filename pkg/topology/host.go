package topology

import (
	"fmt"
	"time"

	"github.com/ai4networks/net4me/pkg/node"
	"github.com/ai4networks/net4me/pkg/port"
	"github.com/neaas/neslink"
	"github.com/segmentio/ksuid"
)

// NewHost creates a new topology host and its underlying node. For this, the
// manager for the provided device is used to create the node. If the manager
// does not exist an error is returned. The host is given an auto generated
// unique ID to couple with the node info. If the node can not be added (e.g.
// name can not be determined), an error will be returned.
func NewHost(device, name string, labels map[string]string, config map[string]any) (*Host, error) {
	addedAt := time.Now()
	manager := node.Device(device)
	if manager == nil {
		return nil, fmt.Errorf("device manager not found: %s", device)
	}
	n, err := manager.Add(name, labels, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create node for host: %w", err)
	}
	host, err := NewHostFromNode(n)
	host.addedAt = addedAt
	return host, err
}

// NewHostFromNode creates a new topology host from a given node. The node can
// be of any device type. The host is given an auto generated unique ID to
// couple with the node info. If the node can not be added (e.g. name can not be
// determined), an error will be returned.
func NewHostFromNode(n node.Node) (*Host, error) {
	id, err := ksuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique host id: %w", err)
	}
	name, err := n.Name()
	if err != nil {
		return nil, fmt.Errorf("failed to get host name from node: %w", err)
	}
	h := &Host{
		id:        id.String(),
		name:      name,
		addedAt:   time.Now(),
		updatedAt: time.Now(),
		labels:    make(map[string]string), //TODO: add labels to node
		topology:  topology,
		node:      n,
	}
	topology.hosts = append(topology.hosts, h)
	return h, nil
}

// Remove removes the underlying node from the host. If the node can not be
// removed, an error will be returned.
func (h *Host) Remove() error {
	manager := node.Device(h.node.Device())
	if manager == nil {
		return fmt.Errorf("device manager not found: %s", h.node.Device())
	}
	if err := manager.Remove(h.node); err != nil {
		return fmt.Errorf("failed to remove node: %w", err)
	}
	for i, host := range topology.hosts {
		if host.id == h.id {
			topology.hosts = append(topology.hosts[:i], topology.hosts[i+1:]...)
			break
		}
	}
	return nil
}

// NodeID returns the ID of the underlying node. Since each device type is
// managed using custom managers, this ID comes with no guarantees of
// uniqueness. However, it should be unique for a given device type.
func (h *Host) NodeID() string {
	return h.node.ID()
}

// NodeInfo returns the information about the underlying node. This information
// is device specific and may vary between device types. If the information can
// not be determined, an error will be returned.
func (h *Host) NodeInfo() (map[string]any, error) {
	return h.node.Info()
}

// Device returns the device type of the underlying node.
func (h *Host) Device() string {
	return h.node.Device()
}

// Topology returns the topology that the host is part of.
func (h *Host) Topology() *Topology {
	return h.topology
}

// NetworkNamespace returns a provider the network namespace of the underlying
// node.
func (h *Host) NetworkNamespace() neslink.NsProvider {
	return h.node.NetNs()
}

// Start starts the underlying node. If the node can not be started, an error
// will be returned.
func (h *Host) Start() error {
	if h.State() != HostStateReady {
		return fmt.Errorf("host is not in ready state")
	}
	return h.node.Start()
}

// Stop stops the underlying node. If the node can not be stopped, an error will
// be returned.
func (h *Host) Stop() error {
	if h.State() != HostStateRunning {
		return fmt.Errorf("host is not in running state")
	}
	return h.node.Stop()
}

// Ports returns the ports of the underlying node. If the node is not in a valid
// state for ports, an error will be returned.
func (h *Host) Ports() ([]port.Port, error) {
	if h.State() != HostStateReady && h.State() != HostStateRunning {
		return nil, fmt.Errorf("host is not in valid state for ports")
	}
	return h.node.Ports()
}

func (h *Host) Stats() (map[string]any, error) {
	if h.State() != HostStateRunning {
		return nil, fmt.Errorf("host is not in running state")
	}
	return h.node.Stats()
}
