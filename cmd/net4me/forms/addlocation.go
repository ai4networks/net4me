package forms

import (
	"fmt"

	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

func AddLocation() {
	siteName := ""
	controllerIP := ""
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name for new location").
				Placeholder("core").
				Prompt("> ").
				CharLimit(7).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("name cannot be empty")
					}
					return nil
				}).
				Value(&siteName),
			huh.NewInput().
				Title("Controller IP for location bridge (e.g. tcp:127.0.0.1:6633)").
				Placeholder("").
				Prompt("> ").
				Validate(func(s string) error {
					return nil
				}).
				Value(&controllerIP),
		),
	)
	if err := form.Run(); err != nil {
		log.Info("add location form was canceled")
		return
	}
	config := map[string]interface{}{
		"controller_ip": controllerIP,
	}
	host, err := topology.NewHost("ovs", siteName, make(map[string]string), config)
	if err != nil {
		log.Error("failed to create location", "error", err.Error())
		return
	} else {
		log.Info("created location", "location", host.Name())
	}
	if err := host.Start(); err != nil {
		log.Error("failed to start site", "error", err.Error())
	} else {
		log.Info("started location", "location", host.Name())
	}
}

func init() {
	addForm("Add Location", AddLocation)
}
