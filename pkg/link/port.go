package link

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

type Port struct {
	index int
	name  string

	nl netlink.Link
}

// PeerID returns the id (link index) of the port paired with the given port.
// For this, the given port must be a veth port.
func (p *Port) PeerID() (int, error) {
	veth, ok := p.nl.(*netlink.Veth)
	if !ok {
		return -1, fmt.Errorf("link is not a veth")
	}
	peerIdx, err := netlink.VethPeerIndex(veth)
	if err != nil {
		return -1, err
	}
	return peerIdx, nil
}
