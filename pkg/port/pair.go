package port

import (
	"fmt"

	"github.com/neaas/neslink"
	"github.com/vishvananda/netlink"
)

func PortPeerIndex(nsP neslink.NsProvider, p Port) (int, error) {
	var peer int
	if err := neslink.Do(
		nsP,
		neslink.LAGeneric("get-peer-index", func() error {
			peerIdx, err := netlink.VethPeerIndex(p.(*netlink.Veth))
			if err != nil {
				return err
			}
			peer = peerIdx
			return nil
		}),
	); err != nil {
		return 0, fmt.Errorf("could not get link peer: %w", err)
	}
	return peer, nil
}

func IsPair(nspA, nspB neslink.NsProvider, a, b Port) (bool, error) {
	var peerA int
	if err := neslink.Do(
		nspA,
		neslink.LAGeneric("get-peer-index", func() error {
			peer, err := netlink.VethPeerIndex(a.(*netlink.Veth))
			if err != nil {
				return err
			}
			peerA = peer
			return nil
		}),
	); err != nil {
		return false, fmt.Errorf("could not get link peer: %w", err)
	}
	if peerA == b.Attrs().Index {
		return true, nil
	}
	return false, nil
}
