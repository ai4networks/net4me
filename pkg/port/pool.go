package port

import (
	"errors"
	"fmt"
	"os"

	"github.com/neaas/neslink"
	"github.com/vishvananda/netlink"
)

// NetNs returns the provider to the network namespace used for port management.
// If the namespace does not exist and/or can not be created, an error will be
// returned.
func NetNs() (neslink.NsProvider, error) {
	netns, _ := neslink.NPNameAt(os.TempDir(), "net4me-ports").Provide()
	if _, err := os.Stat(netns.String()); errors.Is(err, os.ErrNotExist) {
		if err := neslink.Do(
			neslink.NPProcess(os.Getpid()),
			neslink.NANewNsAt(os.TempDir(), "net4me-ports"),
		); err != nil {
			return neslink.NPNameAt(os.TempDir(), "net4me-ports"), fmt.Errorf("could not create port mgmt network namespace: %w", err)
		}
	}
	return neslink.NPNameAt(os.TempDir(), "net4me-ports"), nil
}

// PortPool returns a list of ports found in the port mgmt network namespace. If
// the ports can not be found, an error will be returned. The list of ports can
// be filtered by the provided filters.
func PortPool(filters ...PortFilter) ([]Port, error) {
	netns, err := NetNs()
	if err != nil {
		return nil, fmt.Errorf("could not get port mgmt network namespace: %w", err)
	}
	var links []netlink.Link
	if err := neslink.Do(
		netns,
		neslink.NALinks(&links),
	); err != nil {
		return nil, err
	}
	for _, filter := range filters {
		links = filter(links)
	}
	ports := make([]Port, 0)
	for _, l := range links {
		if l.Attrs().Name == "lo" {
			continue
		}
		if l.Type() != "veth" {
			continue //TODO: review if we need to also include more types of link
		}
		ports = append(ports, Port(l))
	}
	return ports, nil
}

// CreatePortPair creates a new pair of ports in the port mgmt network
// namespace. Each is given a name in the form vp<random5chars>. If the ports
// can not be created, an error will be returned.
func CreatePortPair() (Port, Port, error) {
	netns, err := NetNs()
	if err != nil {
		return nil, nil, err
	}
	port1Name := portNameGenerate()
	port2Name := portNameGenerate()
	var port1, port2 netlink.Link
	if err := neslink.Do(
		netns,
		neslink.LANewVeth(port1Name, port2Name),
		neslink.NAGetLink(neslink.LPName(port1Name), &port1),
		neslink.NAGetLink(neslink.LPName(port2Name), &port2),
	); err != nil {
		return nil, nil, err
	}
	return Port(port1), Port(port2), nil
}

// DestroyPortPair removes the provided port from the port mgmt network
// namespace. If the port can not be removed, an error will be returned. If the
// port is removed without issue, any peer ports will also be removed,
// regardless of their namespace.
func DestroyPortPair(port Port) error {
	netns, err := NetNs()
	if err != nil {
		return err
	}
	if err := neslink.Do(
		netns,
		neslink.LADelete(neslink.LPIndex(port.Attrs().Index)),
	); err != nil {
		return err
	}
	return nil
}

// GivePort moves a provided port back to the pool network namespace. If the
// port can not be moved, an error will be returned. The namespace that the port
// currently resides in must be provided.
func GivePort(nsp neslink.NsProvider, p Port) error {
	netns, err := NetNs()
	if err != nil {
		return err
	}
	if err := neslink.Do(
		nsp,
		neslink.NASetLinkNs(neslink.LPIndex(p.Attrs().Index), netns),
	); err != nil {
		return err
	}
	return nil
}

// TakePort moves a provided port from the pool network namespace to the
// provided namespace. If the port can not be moved, an error will be returned.
func TakePort(nsp neslink.NsProvider, p Port) error {
	netns, err := NetNs()
	if err != nil {
		return err
	}
	if err := neslink.Do(
		netns,
		neslink.NASetLinkNs(neslink.LPIndex(p.Attrs().Index), nsp),
	); err != nil {
		return err
	}
	return nil
}

// ClearPool removes all ports from the port mgmt network namespace. If the
// ports can not be removed, an error will be returned. Note, the only ports
// removed are the ones that are returned by the PortPool function.
func ClearPool() error {
	for {
		poolPorts, err := PortPool()
		if err != nil {
			return err
		}
		if len(poolPorts) == 0 {
			break
		}
		if err := DestroyPortPair(poolPorts[0]); err != nil {
			return err
		}
	}
	return nil
}
