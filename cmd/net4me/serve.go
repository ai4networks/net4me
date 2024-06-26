package main

import (
	"github.com/ai4networks/net4me/pkg/api/control"
	"github.com/ai4networks/net4me/pkg/node"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "serve starts the net4me server",
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
			logrus.Debugln("executing serve command")
			if err := control.Serve(":8080"); err != nil {
				logrus.WithError(err).Fatalln("api server failed")
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
}
