package node

import (
	"github.com/ai4networks/net4me/pkg/port"
	"github.com/neaas/neslink"
)

type Node interface {
	// ID returns the unique identifier for the node. This is **not** guaranteed
	// to be unique for a given topology. However, this ID is unique for the node
	// of a given device type unless otherwise specified.
	ID() string

	// Manager returns the manager that can control the node. This manager is
	// responsible for the lifecycle of the node.
	Manager() Manager

	// Name returns the friendly name for the node. This is not guaranteed to be
	// unique for a given topology, and the true source of the name can differ
	// from device type to device type. If the name can not be determined, an
	// error will be returned.
	Name() (string, error)

	// Device returns the name of the device type for the node. For example, a
	// DinD site or OvS switch.
	Device() string

	// Start starts the node. This assumes that the node has been created and is
	// not already started. If already started, this will not return an error.
	// However, if the node is not created or is stopped and the attempt to start
	// fails, this will return an error.
	Start() error

	// Stop stops the node. This assumes that the node has been created and is not
	// already stopped. If already stopped, this will not return an error.
	// However, if the node is not created or is started and the attempt to stop
	// fails, this will return an error.
	Stop() error

	// Running returns true if the node is running. If the state of the node can
	// not be determined, this will return false.
	Running() bool

	// Info returns a map of information about the node. The keys are the names of
	// the information. Each device type may result in different information
	// being available. If the information can not be determined, an error will
	// be returned.
	Info() (map[string]any, error)

	// NetNs returns a NetNs provider for the node. This provider can be used to
	// interact with the network namespace of the node.
	NetNs() neslink.NsProvider

	// Ports provides a list of port IDs found within a node. If this list can not
	// be determined, an error will be returned.
	Ports() ([]port.Port, error)

	// PortAdd adds a port to the node. The port must already exist and be
	// findable by the node in the port pool. If the port is not added to the
	// node, this will return an error. Note that a node does not need to be
	// started to consume ports.
	PortAdd(port.Port) error

	// PortRemove attempts to remove a port from the node. If the port is not
	// found or can not be removed, this will return an error. Note that a node
	// does not need to be stopped to remove ports. The port is given back to the
	// port pool.
	PortRemove(port.Port) error

	// Stats returns a map of statistics for the node. The keys are the names of
	// the statistics. Each device type may result in different statistics being
	// available. If the statistics can not be determined, an error will be
	// returned.
	Stats() (map[string]any, error)
}

type NodeFilter func([]Node) []Node

func FilterByName(names ...string) NodeFilter {
	return func(nodes []Node) []Node {
		filtered := make([]Node, 0)
		for _, n := range nodes {
			for _, name := range names {
				nodeName, err := n.Name()
				if err != nil {
					continue
				}
				if nodeName == name {
					filtered = append(filtered, n)
				}
			}
		}
		return filtered
	}
}

func FilterByDevice(devices ...string) NodeFilter {
	return func(nodes []Node) []Node {
		filtered := make([]Node, 0)
		for _, n := range nodes {
			for _, device := range devices {
				if n.Device() == device {
					filtered = append(filtered, n)
				}
			}
		}
		return filtered
	}
}

func FilterByID(ids ...string) NodeFilter {
	return func(nodes []Node) []Node {
		filtered := make([]Node, 0)
		for _, n := range nodes {
			for _, id := range ids {
				if n.ID() == id {
					filtered = append(filtered, n)
				}
			}
		}
		return filtered
	}
}
