package main

import (
	"os"

	"github.com/ai4networks/net4me/pkg/node"
	"github.com/ai4networks/net4me/pkg/port"
	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var hosttestCmd = &cobra.Command{
	Use:   "ht",
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

		// get all hosts
		hosts := topology.Hosts()
		for _, host := range hosts {
			logrus.
				WithField("name", host.Name()).
				WithField("id", host.ID()).
				WithField("node_id", host.NodeID()).
				WithField("node_device", host.Device()).
				Infoln("found host")
		}

		os.Exit(0)

		host, err := topology.NewHost("dind", "test", make(map[string]string), make(map[string]any))
		if err != nil {
			logrus.WithError(err).Fatalln("failed to create host")
		}
		logrus.
			WithField("name", host.Name()).
			WithField("id", host.ID()).
			WithField("node_id", host.NodeID()).
			Infoln("created host")

		// get state of host
		logrus.WithField("state", host.State()).Infoln("got host state")

		// get host info
		if info, err := host.NodeInfo(); err != nil {
			logrus.WithError(err).Errorln("failed to get host info")
		} else {
			logrus.WithField("info", info).Infoln("got host info")
		}

		// start host
		if err := host.Start(); err != nil {
			logrus.WithError(err).Fatalln("failed to start host")
		}
		logrus.Infoln("started host")

		// get state of host
		logrus.WithField("state", host.State()).Infoln("got host state")

		// stop host
		if err := host.Stop(); err != nil {
			logrus.WithError(err).Errorln("failed to stop host")
		}
		logrus.Infoln("stopped host")

		// get state of host
		logrus.WithField("state", host.State()).Infoln("got host state")

		// remove host
		if err := host.Remove(); err != nil {
			logrus.WithError(err).Fatalln("failed to remove host")
		}
		logrus.Infoln("removed host")

		// get state of host
		logrus.WithField("state", host.State()).Infoln("got host state")

		// done
		port.ClearPool()
		os.Exit(0)

	},
}

func init() {
	rootCmd.AddCommand(hosttestCmd)
}
