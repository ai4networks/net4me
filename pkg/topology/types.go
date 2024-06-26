package topology

import (
	"time"

	"github.com/ai4networks/net4me/pkg/node"
	"github.com/ai4networks/net4me/pkg/port"
)

type Host struct {
	id        string
	name      string
	addedAt   time.Time
	updatedAt time.Time
	labels    map[string]string

	node     node.Node
	topology *Topology
}

type HostState string

type Link struct {
	self struct {
		host *Host
		port port.Port
	}
	peer struct {
		host *Host
		port port.Port
	}
}

func (h *Host) ID() string {
	return h.id
}

func (h *Host) Name() string {
	return h.name
}

func (h *Host) AddedAt() time.Time {
	return h.addedAt
}

func (h *Host) UpdatedAt() time.Time {
	return h.updatedAt
}

func (h *Host) Labels() map[string]string {
	return h.labels
}

func (h *Host) Node() node.Node {
	return h.node
}

func (l *Link) SelfHost() *Host {
	return l.self.host
}

func (l *Link) SelfPort() port.Port {
	return l.self.port
}

func (l *Link) PeerHost() *Host {
	return l.peer.host
}

func (l *Link) PeerPort() port.Port {
	return l.peer.port
}
