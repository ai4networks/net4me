package forms

import (
	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

func LoadTopology() {
	loadHosts := false
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Load topology from system?").
				Affirmative("Yes (overwrite)").
				Negative("No").
				Value(&loadHosts),
		),
	)
	if err := form.Run(); err != nil {
		log.Info("load confirm form was canceled")
		return
	}
	if err := topology.LoadTopology(); err != nil {
		log.Error("failed to load topology", "error", err.Error())
		return
	}
	log.Info("loaded topology from system", "hosts", len(topology.Hosts()))
}

func init() {
	addForm("Load Topology", LoadTopology)
}
