package forms

import (
	"fmt"
	"net/netip"

	"github.com/ai4networks/net4me/pkg/port"
	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

func AddSite() {
	locations := topology.Hosts()
	if len(locations) == 0 {
		log.Info("no locations to add site to")
		return
	}
	filteredLocations := make([]*topology.Host, 0)
	for _, location := range locations {
		if location.Device() == "ovs" {
			if location.Node().Running() {
				filteredLocations = append(filteredLocations, location)
			}
		}
	}
	locations = filteredLocations
	locationOptions := make([]huh.Option[string], 0, len(locations))
	for _, l := range locations {
		locationOptions = append(locationOptions, huh.NewOption(l.Name(), l.ID()))
	}

	siteName := ""
	locationID := ""
	address := ""
	gpu := false
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name for new site").
				Placeholder("site-1").
				Prompt("> ").
				CharLimit(10).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("name cannot be empty")
					}
					return nil
				}).
				Value(&siteName),
			huh.NewSelect[string]().
				Title("Select location for site").
				Options(locationOptions...).
				Value(&locationID),
			huh.NewInput().
				Title("IPv4 address (cidr) for site").
				Placeholder("10.0.0..../24").
				Prompt("> ").
				CharLimit(18).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("ipv4 cannot be empty")
					}
					if _, err := netip.ParsePrefix(s); err != nil {
						return fmt.Errorf("invalid cidr ipv4 prefix")
					}
					return nil
				}).
				Value(&address),
			huh.NewConfirm().
				Title("Attach GPU to site?").
				Affirmative("Yes").
				Negative("No").
				Value(&gpu),
		),
	)
	if err := form.Run(); err != nil {
		log.Info("add site form was canceled")
		return
	}
	config := map[string]interface{}{
		"gpu": gpu,
	}
	host, err := topology.NewHost("dind", siteName, make(map[string]string), config)
	if err != nil {
		log.Error("failed to create site", "error", err.Error())
		return
	} else {
		log.Info("created site", "site", host.Name())
	}

	for _, l := range locations {
		if l.ID() == locationID {
			lk, err := host.Link(l)
			if err != nil {
				log.Error("failed to link site to location", "error", err.Error())
				if err := host.Remove(); err != nil {
					log.Error("failed to remove site", "error", err.Error())
				}
				return
			} else {
				log.Info("linked site to location", "site", host.Name(), "location", l.Name())
			}
			if err := port.PortAddAddress(lk.SelfHost().NetworkNamespace(), lk.SelfPort(), address); err != nil {
				log.Error("failed to add address to site", "error", err.Error())
			}
			if err := port.PortSetUp(lk.SelfHost().NetworkNamespace(), lk.SelfPort()); err != nil {
				log.Error("failed to set up port on site", "error", err.Error())
			}
			if err := port.PortSetUp(lk.PeerHost().NetworkNamespace(), lk.PeerPort()); err != nil {
				log.Error("failed to set up port on location", "error", err.Error())
			}
		}
	}

	if err := host.Start(); err != nil {
		log.Error("failed to start site", "error", err.Error())
	} else {
		log.Info("started site", "site", host.Name())
	}
}

func init() {
	addForm("Add Site", AddSite)
}
