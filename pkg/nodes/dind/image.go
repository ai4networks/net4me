package dind

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// DinD defines what container image is used to create the docker-in-docker site
// environments. This is at least a combination of the image name (+tag) along
// with the command to run in the container to start the docker engine. The
// image must run a docker engine when the command is run and stop it when the
// command is killed. The docker engine must also be controllable via a unix
// socket located at `/var/run/docker.sock` in the container's filesystem.
type DinD struct {
	imageName  string
	command    []string
	alwaysPull bool
}

// NewDinD creates a new docker-in-docker configuration. This is a combination
// of and image name (+tag) and a command to run when the container is started.
// Any pre-existing entrypoints/commands bundled with the image will be ignored.
func NewDinD(imageName string, command []string, alwaysPull bool) *DinD {
	return &DinD{
		imageName:  imageName,
		command:    command,
		alwaysPull: alwaysPull,
	}
}

// ImageName returns the name of the docker-in-docker image.
func (d *DinD) ImageName() string {
	return d.imageName
}

// Command returns the command to run when the docker-in-docker container is
// started. This command must start the docker engine and expose the docker
// engine via a unix socket at `/var/run/docker.sock`. The command should stay
// attached to the docker engine process and stop the docker engine when the
// command is killed.
func (d *DinD) Command() []string {
	return d.command
}

// ---

func (m *Manager) PullDinD() error {
	if m.dind.ImageName() == "" {
		return fmt.Errorf("site image name is required")
	}
	if !m.siteImageExists(m.dind.imageName) || m.dind.alwaysPull {
		if err := m.siteImagePull(m.dind.imageName); err != nil {
			return fmt.Errorf("failed to pull site image: %w", err)
		}
	}
	id := m.siteImageID(m.dind.imageName)
	if id == "" {
		return fmt.Errorf("failed to get site image id")
	}
	return nil

}

func (m *Manager) siteImagePull(name string) error {
	resp, err := m.clientDocker.ImagePull(context.Background(), name, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull site image: %w", err)
	}
	defer resp.Close()
	io.Copy(io.Discard, resp)
	// TODO: Report any errors that occur during the pull operation.
	return nil
}

func (m *Manager) siteImageExists(name string) bool {
	_, _, err := m.clientDocker.ImageInspectWithRaw(context.Background(), name)
	if err == nil {
		return true
	}
	if client.IsErrNotFound(err) {
		return false
	}
	return false
}

func (m *Manager) siteImageID(name string) string {
	images, err := m.clientDocker.ImageList(context.Background(), image.ListOptions{})
	// TODO: Use filters to reduce the workload of this
	if err != nil {
		return ""
	}
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == name {
				return img.ID
			}
		}
	}
	return ""
}
