package forms

import (
	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

func RemoveHosts() {
	var removeHostNames []string
	currentHosts := topology.Hosts()
	if len(currentHosts) == 0 {
		log.Info("no hosts to remove")
		return
	}
	options := make([]huh.Option[string], 0, len(currentHosts))
	for _, h := range currentHosts {
		options = append(options, huh.NewOption(h.Name(), h.ID()))
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select Hosts to Remove").
				Options(options...).
				Value(&removeHostNames),
		),
	)
	if err := form.Run(); err != nil {
		log.Info("host selection form was canceled")
		return
	}
	for _, h := range currentHosts {
		for _, n := range removeHostNames {
			if h.ID() == n {
				if err := h.Remove(); err != nil {
					log.Error("failed to remove host", "error", err.Error(), "host", h.Name())
				} else {
					log.Info("removed host", "host", h.Name())
				}
			}
		}
	}
}

func init() {
	addForm("Remove Hosts", RemoveHosts)
}
