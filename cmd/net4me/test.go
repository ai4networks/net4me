package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/ai4networks/net4me/pkg/node"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ai4networks/net4me/pkg/port"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "for small unit tests in development",
	Run: func(cmd *cobra.Command, args []string) {
		managers := node.Managers()
		for _, m := range managers {
			logrus.WithField("device", m.Device()).Infoln("found manager")
			if err := m.Setup(nil); err != nil {
				logrus.WithError(err).Fatalln("failed to setup manager")
			}
		}

		// docker
		var dindManager node.Manager
		for _, m := range managers {
			if m.Device() == "dind" {
				dindManager = m
				break
			}
		}

		// try to add
		if n, err := dindManager.Add(
			"test-dind",
			make(map[string]string),
			map[string]any{
				"name": "test-dind",
			}); err != nil {
			logrus.WithError(err).Fatalln("failed to add dind node")
		} else {
			logrus.WithField("id", n.ID()).Infoln("added dind node")
		}

		// get nodes
		nodes, err := dindManager.Nodes(node.FilterByName("test-dind"))
		if err != nil {
			logrus.WithError(err).Fatalln("failed to get nodes")
		}
		logrus.WithField("nodes", nodes).Infoln("got test-dind nodes")

		// create ports
		p1, p2, err := port.CreatePortPair()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to add port pair")
		}
		logrus.WithField("port1", p1.Attrs().Name).Infoln("added port1")
		logrus.WithField("port2", p2.Attrs().Name).Infoln("added port2")

		// add to container
		if err := nodes[0].PortAdd(p1); err != nil {
			logrus.WithError(err).Fatalln("failed to add port to dind node")
		} else {
			logrus.Infoln("added port to dind node")
		}

		// list container ports
		ports, err := nodes[0].Ports()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to get ports")
		}
		for _, p := range ports {
			logrus.WithField("port", p.Attrs().Name).Infoln("found port")
		}

		// get state pre-run
		logrus.WithField("running", nodes[0].Running()).Infoln("got dind node running state")

		if err := nodes[0].Start(); err != nil {
			logrus.WithError(err).Fatalln("failed to start dind node")
		} else {
			logrus.Infoln("started dind node")
		}
		time.Sleep(3 * time.Second)

		// get state post start
		logrus.WithField("running", nodes[0].Running()).Infoln("got dind node running state")

		bufio.NewReader(os.Stdin).ReadBytes('\n')

		// set mac
		if p1, err := port.FromIndex(nodes[0].NetNs(), p1.Attrs().Index); err != nil {
			logrus.WithError(err).Fatalln("failed to get port")
		} else {
			if err := port.PortSetHW(nodes[0].NetNs(), p1, "00:00:00:00:00:01"); err != nil {
				logrus.WithError(err).Fatalln("failed to set port mac")
			}
			logrus.Infoln("set port mac")
		}

		// check mac
		if p1, err := port.FromIndex(nodes[0].NetNs(), p1.Attrs().Index); err != nil {
			logrus.WithError(err).Errorln("failed to get port")
		} else {
			logrus.WithField("mac", p1.Attrs().HardwareAddr.String()).Infoln("got port mac")
		}

		if info, err := nodes[0].Info(); err != nil {
			logrus.WithError(err).Fatalln("failed to get dind node info")
		} else {
			logrus.WithFields(info).Infoln("got dind node info")
		}

		if err := nodes[0].Stop(); err != nil {
			logrus.WithError(err).Fatalln("failed to stop dind node")
		} else {
			logrus.Infoln("stopped dind node")
		}

		if err := dindManager.Remove(nodes[0]); err != nil {
			logrus.WithError(err).Fatalln("failed to remove dind node")
		} else {
			logrus.Infoln("removed dind node")
		}

		port.ClearPool()
		os.Exit(0)

		// ovs
		var ovsManager node.Manager
		for _, m := range managers {
			if m.Device() == "ovs" {
				ovsManager = m
				break
			}
		}
		if ovsManager == nil {
			logrus.WithError(fmt.Errorf("could not find ovs manager")).Fatalln("failed to find ovs manager")
		}

		// try to add
		if n, err := ovsManager.Add(
			"test",
			make(map[string]string),
			map[string]any{
				"name": "test",
			}); err != nil {
			logrus.WithError(err).Fatalln("failed to add ovs node")
		} else {
			logrus.WithField("id", n.ID()).Infoln("added ovs node")
		}

		// link
		poolPorts, err := port.PortPool()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to get port pool")
		}
		logrus.WithField("pool", poolPorts).Infoln("got port pool")

		portNetNs, _ := port.NetNs()
		pair, err := port.IsPair(portNetNs, portNetNs, p1, p2)
		if err != nil {
			logrus.WithError(err).Fatalln("failed to check if pair")
		} else {
			logrus.WithField("pair", pair).Infoln("checked if pair")
		}

		// get nodes
		nodes, err = ovsManager.Nodes(node.FilterByName("test"))
		if err != nil {
			logrus.WithError(err).Fatalln("failed to get nodes")
		}

		poolPorts, err = port.PortPool()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to get port pool")
		}
		logrus.WithField("pool", poolPorts).Infoln("got port pool")
		for _, p := range poolPorts {
			logrus.WithField("mac", p.Attrs().HardwareAddr.String()).Infoln("found port")
		}

		// try to add port
		for _, n := range nodes {
			name, err := n.Name()
			if err != nil {
				logrus.WithError(err).Fatalln("failed to get name")
			}
			if name == "test" {
				if err := n.PortAdd(p1); err != nil {
					logrus.WithError(err).Fatalln("failed to add port to node")
				}
				logrus.Infoln("added port to node")
			}
		}

		poolPorts, err = port.PortPool()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to get port pool")
		}
		logrus.WithField("pool", poolPorts).Infoln("got port pool")
		for _, p := range poolPorts {
			logrus.WithField("mac", p.Attrs().HardwareAddr.String()).Infoln("found port")
		}

		for _, n := range nodes {
			name, err := n.Name()
			if err != nil {
				logrus.WithError(err).Fatalln("failed to get name")
			}
			if name == "test" {
				if err := n.PortRemove(p1); err != nil {
					logrus.WithError(err).Errorln("failed to remove port from node")
				}
				logrus.Infoln("removed port from node")
			}
		}

		poolPorts, err = port.PortPool()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to get port pool")
		}
		logrus.WithField("pool", poolPorts).Infoln("got port pool")
		for _, p := range poolPorts {
			logrus.WithField("mac", p.Attrs().HardwareAddr.String()).Infoln("found port")
		}

		for _, n := range nodes {
			name, err := n.Name()
			if err != nil {
				logrus.WithError(err).Fatalln("failed to get name")
			}
			if name == "test" {
				if err := ovsManager.Remove(n); err != nil {
					logrus.WithError(err).Errorln("failed to remove node")
				}
				logrus.Infoln("removed node")
			}
		}

		port.ClearPool()

	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
