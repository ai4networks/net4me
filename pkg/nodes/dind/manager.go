package dind

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path"
	"sync"

	"github.com/ai4networks/net4me/pkg/node"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/mitchellh/mapstructure"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Manager struct {
	lock         *sync.RWMutex
	dind         *DinD
	socketDir    string
	clientDocker *client.Client
}

type ManagerConfig struct {
	Host       string   `mapstructure:"host"`
	Image      string   `mapstructure:"image"`
	Command    []string `mapstructure:"command"`
	AlwaysPull bool     `mapstructure:"alwaysPull"`
}

type AddConfig struct {
	GPU bool `mapstructure:"gpu"`
}

func (m *Manager) Device() string {
	return "dind"
}

func (m *Manager) Setup(config map[string]any) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	defaultConfig := map[string]any{
		"host":       "unix:///var/run/docker.sock",
		"image":      "docker:dind",
		"command":    []string{"dockerd-entrypoint.sh"},
		"alwaysPull": false,
	}
	maps.Copy(defaultConfig, config)
	c := ManagerConfig{}
	if err := mapstructure.Decode(defaultConfig, &c); err != nil {
		return fmt.Errorf("failed to decode dind manager config: %w", err)
	}
	m.dind = NewDinD(c.Image, c.Command, c.AlwaysPull)
	clientDocker, err := client.NewClientWithOpts(client.WithHost(c.Host), client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	m.clientDocker = clientDocker
	if err := m.PullDinD(); err != nil {
		return fmt.Errorf("failed to find or pull docker-in-docker image: %w", err)
	}
	if path, err := os.MkdirTemp("", "net4me-dind-socket-*"); err != nil {
		return fmt.Errorf("failed to create temporary directory for dind sockets: %w", err)
	} else {
		m.socketDir = path
	}
	return nil
}

func (m *Manager) Info() (map[string]any, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	info, err := m.clientDocker.Info(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get docker info: %w", err)
	}
	return map[string]any{
		"socket_dir": m.socketDir,
		"arch":       info.Architecture,
		"kernel":     info.KernelVersion,
		"version":    info.ServerVersion,
		"os_version": info.OSVersion,
		"driver":     info.Driver,
		"storage":    info.DockerRootDir,
		"memory":     info.MemTotal,
		"cpus":       info.NCPU,
		"labels":     info.Labels,
	}, nil
}

func (m *Manager) Icon() string {
	return "docker"
}

func (m *Manager) Color() string {
	return "RoyalBlue"
}

func (m *Manager) Nodes(nodeFilters ...node.NodeFilter) ([]node.Node, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	nodes := make([]node.Node, 0)
	containers, err := m.clientDocker.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "net4me=true"),
			filters.Arg("label", "net4me.device=dind"),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container list: %w", err)
	}
	for _, container := range containers {
		nodes = append(nodes, m.newNode(container.ID))
	}
	return nodes, nil
}

func (m *Manager) Add(name string, labels map[string]string, config map[string]any) (node.Node, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var c AddConfig
	if err := mapstructure.Decode(config, &c); err != nil {
		return nil, fmt.Errorf("failed to decode docker add config: %w", err)
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	defaultLabels := map[string]string{
		"net4me":             "true",
		"net4me.version":     "v0.0.0", // TODO: get version of net4me application
		"net4me.device":      "dind",
		"net4me.device.name": name,
	}
	maps.Copy(labels, defaultLabels)

	deviceRequests := make([]container.DeviceRequest, 0)
	if c.GPU {
		gpuOpts := &opts.GpuOpts{}
		gpuOpts.Set("all")
		deviceRequests = append(deviceRequests, gpuOpts.Value()...)
	}

	resp, err := m.clientDocker.ContainerCreate(
		context.Background(),
		&container.Config{
			Hostname: name,
			Image:    m.dind.imageName,
			Labels:   labels,
			// NetworkDisabled: true,
			Entrypoint: strslice.StrSlice{"tail"},
			Cmd:        strslice.StrSlice{"-f", "/dev/null"},
		},
		&container.HostConfig{
			AutoRemove: true,
			// NetworkMode: "none",
			Privileged: true,
			Binds: []string{
				path.Join(m.socketDir, name) + ":/var/run/",
				"/:/host",
			},
			Resources: container.Resources{
				DeviceRequests: deviceRequests,
			},
		},
		&network.NetworkingConfig{},
		&v1.Platform{},
		name,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create dind container %s: %w", name, err)
	}
	if err := m.clientDocker.ContainerStart(
		context.Background(),
		resp.ID,
		container.StartOptions{},
	); err != nil {
		return nil, fmt.Errorf("could not start dind container %s: %w", name, err)
	}
	return m.newNode(resp.ID), nil
}

func (m *Manager) Remove(n node.Node) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if err := m.clientDocker.ContainerRemove(context.Background(), n.ID(), container.RemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("failed to remove dind container: %w", err)
	}
	return nil
}

func init() {
	node.RegisterManager(&Manager{
		lock: &sync.RWMutex{},
	})
}
