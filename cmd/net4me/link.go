package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ai4networks/net4me/cmd/net4me/forms"
	"github.com/ai4networks/net4me/pkg/node"
	"github.com/ai4networks/net4me/pkg/port"
	"github.com/ai4networks/net4me/pkg/topology"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "test peer finding in topology",
	PreRun: func(cmd *cobra.Command, args []string) {
		logrus.WithField("manager_count", len(node.Managers())).Debugln("found managers")
		for _, m := range node.Managers() {
			deviceConfig := viper.GetStringMap(m.Device())
			if len(deviceConfig) == 0 {
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

		// topo
		logrus.
			WithField("topology", topology.ID()).
			WithField("manager_count", len(node.Managers())).
			Infoln("created topology")

		// create hosts
		h1, err := topology.NewHost("dind", "host1", make(map[string]string), make(map[string]interface{}))
		if err != nil {
			logrus.WithError(err).Fatalln("failed to create host 1")
		}
		h2, err := topology.NewHost("dind", "host2", make(map[string]string), make(map[string]interface{}))
		if err != nil {
			logrus.WithError(err).Fatalln("failed to create host 2")
		}
		s1, err := topology.NewHost("ovs", "switch1", make(map[string]string), make(map[string]interface{}))
		if err != nil {
			logrus.WithError(err).Fatalln("failed to create switch 1")
		}
		logrus.Infoln("created hosts")

		// start hosts
		if err := h1.Start(); err != nil {
			logrus.WithError(err).Fatalln("failed to start host 1")
		}
		if err := h2.Start(); err != nil {
			logrus.WithError(err).Fatalln("failed to start host 2")
		}
		if err := s1.Start(); err != nil {
			logrus.WithError(err).Fatalln("failed to start switch 1")
		}
		logrus.Infoln("started hosts")

		// link hosts
		if l, err := h1.Link(s1); err != nil {
			logrus.WithError(err).Errorln("failed to link hosts")
		} else {
			logrus.WithField("link", l).Infoln("linked host 1")
			// add addresses
			if err := port.PortAddAddress(l.SelfHost().NetworkNamespace(), l.SelfPort(), "192.168.5.5/24"); err != nil {
				logrus.WithError(err).Errorln("failed to add address to host 1")
			}
			// set up
			if err := port.PortSetUp(l.SelfHost().NetworkNamespace(), l.SelfPort()); err != nil {
				logrus.WithError(err).Errorln("failed to set up port on host 1")
			}
			if err := port.PortSetUp(l.PeerHost().NetworkNamespace(), l.PeerPort()); err != nil {
				logrus.WithError(err).Errorln("failed to set up port on switch 1")
			}
		}

		// link hosts
		if l, err := h2.Link(s1); err != nil {
			logrus.WithError(err).Errorln("failed to link hosts")
		} else {
			logrus.WithField("link", l).Infoln("linked host 2")
			// add addresses
			if err := port.PortAddAddress(l.SelfHost().NetworkNamespace(), l.SelfPort(), "192.168.5.4/24"); err != nil {
				logrus.WithError(err).Errorln("failed to add address to host 2")
			}
			// set up
			if err := port.PortSetUp(l.SelfHost().NetworkNamespace(), l.SelfPort()); err != nil {
				logrus.WithError(err).Errorln("failed to set up port on host 2")
			}
			if err := port.PortSetUp(l.PeerHost().NetworkNamespace(), l.PeerPort()); err != nil {
				logrus.WithError(err).Errorln("failed to set up port on switch 1")
			}
		}

		// list link on host 1
		if links, err := h1.Links(); err != nil {
			logrus.WithError(err).Errorln("failed to list links on host 1")
		} else {
			for _, l := range links {
				logrus.
					WithField("self_port", l.SelfPort().Attrs().Name).
					WithField("peer_port", l.PeerPort().Attrs().Name).
					WithField("self_name", l.SelfHost().Name()).
					WithField("peer_name", l.PeerHost().Name()).
					Infoln("found link on host 1")
			}
		}

		// get stats
		for i := 0; i < 1; i++ {
			if stats, err := h1.Stats(); err != nil {
				logrus.WithError(err).Errorln("failed to get stats for host 1")
			} else {
				logrus.WithFields(stats).Infoln("got stats for host 1")
			}
			time.Sleep(3 * time.Second)
		}

		// connect to host1 engine
		info, err := h1.NodeInfo()
		if err != nil {
			logrus.WithError(err).Errorln("failed to get node info for h1")
		}

		h1dkr, err := client.NewClientWithOpts(client.WithHost(fmt.Sprintf("unix://%s", info["socket"].(string))))
		if err != nil {
			logrus.WithError(err).Errorln("failed to connect to host 1 engine")
		} else {
			h1info, err := h1dkr.Info(context.Background())
			if err != nil {
				logrus.WithError(err).Errorln("failed to get host 1 info")
			} else {
				logrus.WithField("host1", h1info.Name).Infoln("got host 1 info")
			}
		}

		forms.RemoveHosts()

		bufio.NewReader(os.Stdin).ReadBytes('\n')

		// unlink hosts
		if err := h1.Unlink(h2); err != nil {
			logrus.WithError(err).Fatalln("failed to unlink hosts")
		}
		if err := h2.Unlink(s1); err != nil {
			logrus.WithError(err).Fatalln("failed to unlink hosts")
		}
		logrus.Infoln("unlinked hosts")

		// stop hosts
		if err := h1.Stop(); err != nil {
			logrus.WithError(err).Fatalln("failed to stop host 1")
		}
		if err := h2.Stop(); err != nil {
			logrus.WithError(err).Fatalln("failed to stop host 2")
		}
		logrus.Infoln("stopped hosts")

		// destroy hosts
		if err := h1.Remove(); err != nil {
			logrus.WithError(err).Fatalln("failed to destroy host 1")
		}
		if err := h2.Remove(); err != nil {
			logrus.WithError(err).Fatalln("failed to destroy host 2")
		}
		if err := s1.Remove(); err != nil {
			logrus.WithError(err).Fatalln("failed to destroy switch 1")
		}
		logrus.Infoln("destroyed hosts")

	},
}

func init() {
	rootCmd.AddCommand(linkCmd)
}
