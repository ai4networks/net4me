package topology

import "github.com/ai4networks/net4me/pkg/port"

func (h *Host) Link(peer *Host) (*Link, error) {
	sp, pp, err := port.CreatePortPair()
	if err != nil {
		return nil, err
	}
	if err := h.node.PortAdd(sp); err != nil {
		return nil, err
	}
	if err := peer.node.PortAdd(pp); err != nil {
		return nil, err
	}
	return &Link{
		self: struct {
			host *Host
			port port.Port
		}{
			host: h,
			port: sp,
		},
		peer: struct {
			host *Host
			port port.Port
		}{
			host: peer,
			port: pp,
		},
	}, nil
}

func (h *Host) Unlink(peer *Host) error {
	links, err := h.Links()
	if err != nil {
		return err
	}
	for _, l := range links {
		if l.peer.host == peer {
			if err := h.node.PortRemove(l.self.port); err != nil {
				return err
			}
			if err := peer.node.PortRemove(l.peer.port); err != nil {
				return err
			}
			if err := port.DestroyPortPair(l.self.port); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Host) Links() ([]*Link, error) {
	selfPorts, err := h.Ports()
	if err != nil {
		return nil, err
	}
	links := make([]*Link, 0)
	for _, p := range selfPorts {
		peerIdx, err := port.PortPeerIndex(h.NetworkNamespace(), p)
		if err != nil {
			continue
		}
		for _, ph := range topology.hosts {
			peerPorts, err := ph.Ports()
			if err != nil {
				continue
			}
			for _, pp := range peerPorts {
				if pp.Attrs().Index == peerIdx {
					links = append(links, &Link{
						self: struct {
							host *Host
							port port.Port
						}{
							host: h,
							port: p,
						},
						peer: struct {
							host *Host
							port port.Port
						}{
							host: ph,
							port: pp,
						},
					})
				}
			}
		}
	}
	return links, nil
}
