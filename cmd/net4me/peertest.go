package main

import (
	"os"

	"github.com/ai4networks/net4me/pkg/node"
	"github.com/ai4networks/net4me/pkg/port"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var peertestCmd = &cobra.Command{
	Use:   "pt",
	Short: "test peer finding in topology",
	PreRun: func(cmd *cobra.Command, args []string) {
		logrus.WithField("manager_count", len(node.Managers())).Debugln("found managers")
		for _, m := range node.Managers() {
			deviceConfig := viper.GetStringMap(m.Device())
			if deviceConfig == nil {
				logrus.WithField("device", m.Device()).Warnln("custom device manager configuration not found")
			}
			if err := m.Setup(deviceConfig); err != nil {
				logrus.WithError(err).WithField("device", m.Device()).Errorln("failed to setup manager")
				continue
			}
			logrus.WithField("device", m.Device()).Infoln("setup device manager")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ovsManager := node.Device("ovs")
		dindManager := node.Device("dind")
		if ovsManager == nil || dindManager == nil {
			logrus.Fatalln("failed to find managers")
		}

		// make test bridge
		if _, err := ovsManager.Add("test", make(map[string]string), make(map[string]any)); err != nil {
			logrus.WithError(err).Fatalln("failed to add test bridge")
		}

		// get test bridge
		bridges, err := ovsManager.Nodes(node.FilterByName("test"))
		if err != nil {
			logrus.WithError(err).Fatalln("failed to get test bridge")
		}
		if len(bridges) == 0 {
			logrus.Fatalln("no test bridge found")
		}
		bridge := bridges[0]
		logrus.WithField("bridge-id", bridge.ID()).Infoln("found test bridge")

		// get state state
		logrus.WithField("running", bridge.Running()).Infoln("got bridge node running state")

		if err := bridge.Start(); err != nil {
			logrus.WithError(err).Fatalln("failed to start bridge node")
		}

		// get state state
		logrus.WithField("running", bridge.Running()).Infoln("got bridge node running state post start")

		// create port pair
		p1, p2, err := port.CreatePortPair()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to add port pair")
		}
		logrus.WithField("name", p1.Attrs().Name).WithField("idx", p1.Attrs().Index).Infoln("added port1")
		logrus.WithField("name", p2.Attrs().Name).WithField("idx", p2.Attrs().Index).Infoln("added port2")

		// move port 1 to bridge
		if err := bridge.PortAdd(p1); err != nil {
			logrus.WithError(err).Fatalln("failed to add port to bridge")
		}

		// list bridge ports
		ports, err := bridge.Ports()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to get bridge ports")
		}
		for _, p := range ports {
			logrus.WithField("port", p.Attrs().Name).Infoln("found port")
		}

		// PORT 2
		logrus.WithField("index", p2.Attrs().Index).Infoln("searching for peer of port")
		poolNs, err := port.NetNs()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to get port pool namespace")
		}
		peerIdx, err := port.PortPeerIndex(poolNs, p2)
		logrus.WithField("peer-index", peerIdx).Infoln("found peer index")
		for _, m := range node.Managers() {
			nodes, err := m.Nodes()
			if err != nil {
				logrus.WithError(err).WithField("device", m.Device()).Fatalln("failed to get nodes")
			}
			for _, n := range nodes {
				ports, err := n.Ports()
				if err != nil {
					logrus.WithError(err).WithField("node", n.ID()).Fatalln("failed to get ports")
				}
				for _, p := range ports {
					if p.Attrs().Index == peerIdx {
						name, err := n.Name()
						if err != nil {
							logrus.WithError(err).WithField("node", n.ID()).Errorln("failed to get node name")
						}
						logrus.WithField("node", n.ID()).WithField("name", name).Infoln("found node with peer port")
					}
				}
			}
		}

		// done
		port.ClearPool()
		os.Exit(0)

	},
}

func init() {
	rootCmd.AddCommand(peertestCmd)
}
