package port

import (
	"fmt"

	"github.com/neaas/neslink"
	"github.com/vishvananda/netlink"
)

type Port netlink.Link

func PortSetHW(nsp neslink.NsProvider, port Port, hw string) error {
	return neslink.Do(
		nsp,
		neslink.LASetHw(neslink.LPIndex(port.Attrs().Index), hw),
	)
}

func PortSetName(nsp neslink.NsProvider, port Port, name string) error {
	return neslink.Do(
		nsp,
		neslink.LASetName(neslink.LPIndex(port.Attrs().Index), name),
	)
}

func PortSetUp(nsp neslink.NsProvider, port Port) error {
	return neslink.Do(
		nsp,
		neslink.LASetUp(neslink.LPIndex(port.Attrs().Index)),
	)
}

func PortSetDown(nsp neslink.NsProvider, port Port) error {
	return neslink.Do(
		nsp,
		neslink.LASetDown(neslink.LPIndex(port.Attrs().Index)),
	)
}

func PortAddAddress(nsp neslink.NsProvider, port Port, cidr string) error {
	return neslink.Do(
		nsp,
		neslink.LAAddAddr(neslink.LPIndex(port.Attrs().Index), cidr),
	)
}

func PortDelAddress(nsp neslink.NsProvider, port Port, cidr string) error {
	return neslink.Do(
		nsp,
		neslink.LADelAddr(neslink.LPIndex(port.Attrs().Index), cidr),
	)
}

func PortSetEmulation(nsp neslink.NsProvider, port Port, latency, jitter uint32, loss float32) error {
	return neslink.Do(
		nsp,
		neslink.LAAddNetem(neslink.LPIndex(port.Attrs().Index), latency, jitter, loss),
	)
}

type PortFilter func([]netlink.Link) []netlink.Link

// Ports returns a list of all port indexes in the network namespace.
func Ports(netns neslink.NsProvider, filters ...PortFilter) ([]Port, error) {
	var links []netlink.Link
	if err := neslink.Do(
		netns,
		neslink.NALinks(&links),
	); err != nil {
		return nil, fmt.Errorf("could not get links from namespace: %w", err)
	}
	for _, filter := range filters {
		links = filter(links)
	}
	ports := make([]Port, 0)
	for _, link := range links {
		ports = append(ports, Port(link))
	}
	return ports, nil
}

func FilterHasMasterIndex(index int) PortFilter {
	return func(links []netlink.Link) []netlink.Link {
		filtered := make([]netlink.Link, 0)
		for _, link := range links {
			if link.Attrs().MasterIndex == index {
				filtered = append(filtered, link)
			}
		}
		return filtered
	}
}

func FilterHasNameIn(name ...string) PortFilter {
	return func(links []netlink.Link) []netlink.Link {
		filtered := make([]netlink.Link, 0)
		for _, link := range links {
			for _, n := range name {
				if link.Attrs().Name == n {
					filtered = append(filtered, link)
					break
				}
			}
		}
		return filtered
	}
}

func FilterHasTypeIn(typ ...string) PortFilter {
	return func(links []netlink.Link) []netlink.Link {
		filtered := make([]netlink.Link, 0)
		for _, link := range links {
			for _, t := range typ {
				if link.Type() == t {
					filtered = append(filtered, link)
					break
				}
			}
		}
		return filtered
	}
}

func FromIndex(nsp neslink.NsProvider, index int) (Port, error) {
	var link netlink.Link
	if err := neslink.Do(
		nsp,
		neslink.NAGetLink(neslink.LPIndex(index), &link),
	); err != nil {
		return nil, err
	}
	return link, nil
}

func FromName(nsp neslink.NsProvider, name string) (Port, error) {
	var link netlink.Link
	if err := neslink.Do(
		nsp,
		neslink.NAGetLink(neslink.LPName(name), &link),
	); err != nil {
		return nil, err
	}
	return link, nil
}
