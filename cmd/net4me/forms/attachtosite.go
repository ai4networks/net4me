package forms

import (
	"fmt"

	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/docker/docker/client"
	"github.com/neaas/nescript"
	"github.com/neaas/nescript/docker"
)

func SiteExec() {
	var siteID string
	currentHosts := topology.Hosts()
	if len(currentHosts) == 0 {
		log.Info("no sites to attach to")
		return
	}
	options := make([]huh.Option[string], 0, len(currentHosts))
	for _, h := range currentHosts {
		if h.Device() == "dind" {
			options = append(options, huh.NewOption(h.Name(), h.ID()))
		}
	}
	command := "docker image ls"
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Site").
				Options(options...).
				Value(&siteID),
			huh.NewInput().
				Title("Command").
				Placeholder("docker image ls").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("command cannot be empty")
					}
					return nil
				}).
				Value(&command),
		),
	)
	if err := form.Run(); err != nil {
		log.Info("site selection form was canceled")
		return
	}
	localDocker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Error("failed to connect to local docker engine", "error", err.Error())
	}
	for _, h := range currentHosts {
		if h.ID() == siteID {
			script := nescript.NewScript(command)
			process, err := script.Cmd().Exec(docker.Executor(localDocker, h.NodeID(), ""))
			if err != nil {
				log.Error("failed to execute command", "error", err.Error())
				return
			}
			result, err := process.Result()
			if err != nil {
				log.Error("failed to get result from command", "error", err.Error())
				return
			}
			fmt.Printf("%s\n", result.StdOut)
		}
	}
}

func init() {
	addForm("Execute on Site", SiteExec)
}
