package dind

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ai4networks/net4me/pkg/node"
	"github.com/ai4networks/net4me/pkg/port"
	"github.com/neaas/nescript"
	ds "github.com/neaas/nescript/docker"
	"github.com/neaas/neslink"
	"github.com/neaas/neslink/docker"
)

type Node struct {
	manager *Manager
	id      string
}

func (m *Manager) newNode(containerID string) node.Node {
	return &Node{
		id:      containerID,
		manager: m,
	}
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) Manager() node.Manager {
	return n.manager
}

func (n *Node) Name() (string, error) {
	n.manager.lock.RLock()
	defer n.manager.lock.RUnlock()
	container, err := n.manager.clientDocker.ContainerInspect(context.Background(), n.id)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(container.Name, "/"), nil
}

func (n *Node) Device() string {
	return n.manager.Device()
}

func (n *Node) Start() error {
	n.manager.lock.RLock()
	defer n.manager.lock.RUnlock()
	if len(n.manager.dind.command) == 0 {
		return nil
	}
	var cmd *nescript.Cmd
	if len(n.manager.dind.command) == 1 {
		cmd = nescript.NewCmd(n.manager.dind.command[0])
	} else {
		cmd = nescript.NewCmd(n.manager.dind.command[0], n.manager.dind.command[1:]...)
	}
	if _, err := cmd.Exec(ds.Executor(n.manager.clientDocker, n.id, "")); err != nil {
		return err
	}
	checkCounter := 0
	for {
		if n.Running() {
			break
		}
		checkCounter := checkCounter + 1
		if checkCounter > 10 {
			return fmt.Errorf("docker-in-docker container did not start in time")
		}
		time.Sleep(500)
	}
	return nil
}

func (n *Node) Stop() error {
	n.manager.lock.RLock()
	defer n.manager.lock.RUnlock()
	resp, err := n.manager.clientDocker.ContainerTop(
		context.Background(),
		n.id,
		make([]string, 0),
	)
	if err != nil {
		return err
	}
	if len(resp.Processes) == 0 {
		return nil
	}
	commandIndex := -1
	pidIndex := -1
	for i, title := range resp.Titles {
		if title == "CMD" {
			commandIndex = i
		}
		if title == "PID" {
			pidIndex = i
		}
	}
	if commandIndex == -1 || pidIndex == -1 {
		return fmt.Errorf("could not find docker process in dind")
	}
	for _, process := range resp.Processes {
		if strings.Contains(process[commandIndex], "dockerd --host=unix:///var/run/docker.sock") {
			pid, err := strconv.Atoi(process[pidIndex])
			if err != nil {
				return fmt.Errorf("could not find docker process in dind")
			}
			return syscall.Kill(pid, syscall.SIGTERM)
		}
	}
	return nil
}

func (n *Node) Running() bool {
	resp, err := n.manager.clientDocker.ContainerTop(
		context.Background(),
		n.id,
		make([]string, 0),
	)
	if err != nil {
		return false
	}
	commandIndex := -1
	for i, title := range resp.Titles {
		if title == "CMD" {
			commandIndex = i
			break
		}
	}
	if commandIndex == -1 {
		return false
	}
	for _, process := range resp.Processes {
		if strings.Contains(process[commandIndex], "dockerd --host=unix:///var/run/docker.sock") {
			return true
		}
	}
	return false
}

func (n *Node) Info() (map[string]any, error) {
	n.manager.lock.RLock()
	defer n.manager.lock.RUnlock()
	container, err := n.manager.clientDocker.ContainerInspect(context.Background(), n.id)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":      container.ID,
		"name":    container.Name,
		"created": container.Created,
		"state":   container.State.Status,
		"socket":  path.Join(n.manager.socketDir, container.Name, "docker.sock"),
	}, nil
}

func (n *Node) NetNs() neslink.NsProvider {
	return docker.NPID(n.manager.clientDocker, n.id)
}

func (n *Node) Ports() ([]port.Port, error) {
	n.manager.lock.RLock()
	defer n.manager.lock.RUnlock()
	ports, err := port.Ports(n.NetNs(), port.FilterHasTypeIn("veth"))
	if err != nil {
		return nil, err
	}
	return ports, nil
}

func (n *Node) PortAdd(p port.Port) error {
	n.manager.lock.RLock()
	defer n.manager.lock.RUnlock()
	if err := port.TakePort(n.NetNs(), p); err != nil {
		return err
	}
	return nil
}

func (n *Node) PortRemove(p port.Port) error {
	n.manager.lock.RLock()
	defer n.manager.lock.RUnlock()
	if err := port.GivePort(n.NetNs(), p); err != nil {
		return err
	}
	return nil
}
