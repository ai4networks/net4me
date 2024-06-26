package topology

import "github.com/segmentio/ksuid"

type Topology struct {
	id    string
	hosts []*Host
}

var (
	topology *Topology = new()
)

func new() *Topology {
	return &Topology{
		id:    ksuid.New().String(),
		hosts: make([]*Host, 0),
	}
}

func ID() string {
	return topology.id
}

func GetTopology() *Topology {
	return topology
}

// LoadTopology loads the topology (hosts and links) by querying the node
// managers that have been registered and setup. This will generate new IDs for
// all elements of the topology.
func LoadTopology() error {
	hosts, err := hosts()
	if err != nil {
		return err
	}
	topology.hosts = hosts
	return nil
}

func Hosts() []*Host {
	return topology.hosts
}

func Links() []*Link {
	links := make([]*Link, 0)
	for _, h := range topology.hosts {
		l, err := h.Links()
		if err != nil {
			continue
		}
		for _, hostlink := range l {
			duplicate := false
			for _, link := range links {
				if hostlink.SelfPort().Attrs().Index == link.PeerPort().Attrs().Index &&
					hostlink.PeerPort().Attrs().Index == link.SelfPort().Attrs().Index {
					duplicate = true
				}
			}
			if !duplicate {
				links = append(links, l...)
			}
		}

	}
	return links
}
