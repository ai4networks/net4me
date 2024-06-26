package forms

import (
	"fmt"

	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

func DumpSite() {
	var targetHost string
	currentHosts := topology.Hosts()
	if len(currentHosts) == 0 {
		log.Info("no sites to dump")
		return
	}
	options := make([]huh.Option[string], 0, len(currentHosts))
	for _, h := range currentHosts {
		if h.Device() == "dind" {
			options = append(options, huh.NewOption(h.Name(), h.ID()))
		}
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Site to Dump").
				Options(options...).
				Value(&targetHost),
		),
	)
	if err := form.Run(); err != nil {
		log.Info("host selection form was canceled")
		return
	}
	for _, h := range currentHosts {
		if h.ID() == targetHost {
			info, err := h.Node().Info()
			if err != nil {
				log.Error("failed to get site info", "error", err.Error())
				return
			}
			fmt.Printf("Container ID: %s\n", info["id"])
			fmt.Printf("Docker Socket Path: %s\n", info["socket"])
		}
	}
}

func init() {
	addForm("Get Site Info", DumpSite)
}
