package forms

import (
	"context"
	"strconv"

	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func SetSiteResources() {
	var targetHost string
	var cpusS string
	currentHosts := topology.Hosts()
	if len(currentHosts) == 0 {
		log.Info("no hosts to remove")
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
				Title("Select Host to Update").
				Options(options...).
				Value(&targetHost),
			huh.NewInput().
				Title("CPUs").
				Placeholder("1").
				Prompt("> ").
				Validate(func(s string) error {
					if _, err := strconv.ParseFloat(s, 64); err != nil {
						return err
					}
					return nil
				}).
				Value(&cpusS),
		),
	)
	if err := form.Run(); err != nil {
		log.Info("host selection form was canceled")
		return
	}
	cpus, _ := strconv.ParseFloat(cpusS, 64)
	localDocker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Error("failed to connect to local docker engine", "error", err.Error())
	}
	for _, h := range currentHosts {
		if h.ID() == targetHost {
			localDocker.ContainerUpdate(context.Background(), h.NodeID(), container.UpdateConfig{
				Resources: container.Resources{
					CPUQuota: int64(cpus * 100000),
				},
			})
		}
	}
}

func init() {
	addForm("Set Site Resources", SetSiteResources)
}
