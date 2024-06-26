package forms

import (
	"strconv"

	"github.com/ai4networks/net4me/pkg/port"
	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

func ConnectLocations() {
	var locationSrc, locationDst string

	currentHosts := topology.Hosts()
	filteredHosts := make([]*topology.Host, 0)
	for _, h := range currentHosts {
		if h.Device() == "ovs" {
			filteredHosts = append(filteredHosts, h)
		}
	}
	currentHosts = filteredHosts
	if len(currentHosts) < 2 {
		log.Info("need at least 2 locations for this action")
		return
	}
	optionsSrc := make([]huh.Option[string], 0)
	for _, h := range currentHosts {
		optionsSrc = append(optionsSrc, huh.NewOption(h.Name(), h.ID()))
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select the first location").
				Options(optionsSrc...).
				Value(&locationSrc),
		),
	)
	if err := form.Run(); err != nil {
		log.Info("location 1 selection form was canceled")
		return
	}

	optionsDst := make([]huh.Option[string], 0)
	for _, h := range currentHosts {
		if h.ID() != locationSrc {
			optionsDst = append(optionsDst, huh.NewOption(h.Name(), h.ID()))
		}
	}
	formDst := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select the second location").
				Options(optionsDst...).
				Value(&locationDst),
		),
	)
	if err := formDst.Run(); err != nil {
		log.Info("location 2 selection form was canceled")
		return
	}

	var latencyS, jitterS, lossS string
	formEm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("How many microseconds of latency to apply to link?").
				Validate(func(s string) error {
					_, err := strconv.ParseInt(s, 10, 0)
					return err
				}).
				Value(&latencyS),
			huh.NewInput().
				Title("How many microseconds of jitter to apply to link?").
				Validate(func(s string) error {
					_, err := strconv.ParseInt(s, 10, 0)
					return err
				}).
				Value(&jitterS),
			huh.NewInput().
				Title("What percent of packets should be lost on the link").
				Validate(func(s string) error {
					_, err := strconv.ParseFloat(s, 32)
					return err
				}).
				Value(&lossS),
		),
	)
	if err := formEm.Run(); err != nil {
		log.Info("emualtion config form was canceled")
		return
	}

	var src, dst *topology.Host
	for _, h := range currentHosts {
		if h.ID() == locationSrc {
			src = h
		}
		if h.ID() == locationDst {
			dst = h
		}
	}
	if src == nil || dst == nil {
		log.Error("failed to find locations")
		return
	}
	link, err := src.Link(dst)
	if err != nil {
		log.Error("failed to connect locations", "error", err.Error())
		return
	}
	if err := port.PortSetUp(link.SelfHost().NetworkNamespace(), link.SelfPort()); err != nil {
		log.Error("failed to set up port", "error", err.Error())
		return
	}
	if err := port.PortSetUp(link.PeerHost().NetworkNamespace(), link.PeerPort()); err != nil {
		log.Error("failed to set up port", "error", err.Error())
		return
	}
	log.Info("connected locations", "src", src.Name(), "dst", dst.Name())
	latency, errL := strconv.ParseInt(latencyS, 10, 0)
	jitter, errJ := strconv.ParseInt(jitterS, 10, 0)
	loss, errP := strconv.ParseFloat(lossS, 32)
	if errL != nil || errJ != nil || errP != nil {
		log.Info("skipping emulation setup")
		return
	}
	if err := port.PortSetEmulation(link.SelfHost().NetworkNamespace(), link.SelfPort(), uint32(latency), uint32(jitter), float32(loss)); err != nil {
		log.Error("failed to set emulation", "error", err.Error())
		return

	}
	if err := port.PortSetEmulation(link.PeerHost().NetworkNamespace(), link.PeerPort(), uint32(latency), uint32(jitter), float32(loss)); err != nil {
		log.Error("failed to set emulation", "error", err.Error())
		return
	}

}

func init() {
	addForm("Connect Locations", ConnectLocations)
}
