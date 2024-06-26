package forms

import (
	"context"
	"fmt"

	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func AddImageToSite() {
	var targetSite string
	currentHosts := topology.Hosts()
	if len(currentHosts) == 0 {
		log.Warn("no sites to add images too")
		return
	}
	siteOptions := make([]huh.Option[string], 0)
	for _, h := range currentHosts {
		if h.Device() == "dind" {
			siteOptions = append(siteOptions, huh.NewOption(h.Name(), h.ID()))
		}
	}
	var targetImages []string
	localDocker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Error("failed to connect to local docker engine", "error", err.Error())
	}
	imagesList, err := localDocker.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		log.Error("failed to list images from local docker engine", "error", err.Error())
	}
	imageOptions := make([]huh.Option[string], 0)
	for _, i := range imagesList {
		var tag string
		if len(i.RepoTags) > 0 {
			tag = i.RepoTags[0]
		} else {
			continue
		}
		imageOptions = append(imageOptions, huh.NewOption(tag, i.ID))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select target site for image").
				Options(siteOptions...).
				Value(&targetSite),

			huh.NewMultiSelect[string]().
				Title("Select images to transfer").
				Options(imageOptions...).
				Value(&targetImages),
		),
	)
	if err := form.Run(); err != nil {
		log.Info("image tansfer form was canceled")
		return
	}

	var siteDocker *client.Client
	for _, h := range currentHosts {
		if h.ID() == targetSite {
			siteInfo, err := h.Node().Info()
			if err != nil {
				log.Error("failed to get site info", "error", err.Error())
				return
			}
			if socket, ok := siteInfo["socket"]; !ok {
				log.Error("site does not have a known docker socket")
				return
			} else {
				log.Info("connecting to site docker engine", "socket", socket.(string))
				sD, err := client.NewClientWithOpts(client.WithHost(fmt.Sprintf("unix://%s", socket.(string))), client.WithAPIVersionNegotiation())
				if err != nil {
					log.Error("failed to connect to site docker engine", "error", err.Error())
					return
				} else {
					log.Info("connected to site docker engine")
				}
				siteDocker = sD
			}
		}
	}

	for _, i := range targetImages {
		reader, err := localDocker.ImageSave(context.Background(), []string{i})
		if err != nil {
			log.Error("failed to save image", "error", err.Error())
			return
		}
		if _, err := siteDocker.ImageLoad(context.Background(), reader, true); err != nil {
			log.Error("failed to load image", "error", err.Error())
			return
		}
		tag := ""
		for _, img := range imagesList {
			if img.ID == i && len(img.RepoTags) > 0 {
				tag = img.RepoTags[0]
			}
		}
		if err := siteDocker.ImageTag(context.Background(), i, tag); err != nil {
			log.Warn("failed to tag image on site", "error", err.Error())
		}
		log.Info("transferred image to site", "image", i)
	}

}

func init() {
	addForm("Add Image To Site", AddImageToSite)
}
