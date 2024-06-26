package topology

import (
	"github.com/ai4networks/net4me/pkg/node"
)

func hosts() ([]*Host, error) {
	var hosts []*Host
	for _, manager := range node.Managers() {
		nodes, err := manager.Nodes()
		if err != nil {
			return nil, err
		}
		for _, n := range nodes {
			host, err := NewHostFromNode(n)
			if err != nil {
				return nil, err
			}
			hosts = append(hosts, host)
		}
	}
	return hosts, nil
}

type HostFilter func([]*Host) []*Host

func FilterByName(names ...string) HostFilter {
	return func(hosts []*Host) []*Host {
		filtered := make([]*Host, 0)
		for _, n := range hosts {
			for _, name := range names {
				if n.Name() == name {
					filtered = append(filtered, n)
				}
			}
		}
		return filtered
	}
}

func FilterByDevice(devices ...string) HostFilter {
	return func(hosts []*Host) []*Host {
		filtered := make([]*Host, 0)
		for _, n := range hosts {
			for _, device := range devices {
				if n.Device() == device {
					filtered = append(filtered, n)
				}
			}
		}
		return filtered
	}
}

func FilterByID(ids ...string) HostFilter {
	return func(hosts []*Host) []*Host {
		filtered := make([]*Host, 0)
		for _, n := range hosts {
			for _, id := range ids {
				if n.ID() == id {
					filtered = append(filtered, n)
				}
			}
		}
		return filtered
	}
}
