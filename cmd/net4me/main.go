package main

import (
	"fmt"

	"github.com/ai4networks/net4me/cmd/net4me/forms"
	"github.com/ai4networks/net4me/pkg/influx"
	"github.com/ai4networks/net4me/pkg/node"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/ai4networks/net4me/pkg/nodes/dind"
	_ "github.com/ai4networks/net4me/pkg/nodes/ovs"
)

var (
	rootCmd = &cobra.Command{
		Use:   "net4me",
		Short: "net4me is a network emulation tool",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if err := viper.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					logrus.WithError(err).Fatalln("failed to read configuration file")
				}
			}
			if viper.GetBool("verbose") {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				logrus.SetLevel(logrus.InfoLevel)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			log.Info("starting net4me")
			for _, m := range node.Managers() {
				deviceConfig := viper.GetStringMap(fmt.Sprintf("manager.%s", m.Device()))
				if len(deviceConfig) == 0 {
					log.Warn("custom device manager configuration not found", "device", m.Device())
				}
				if err := m.Setup(deviceConfig); err != nil {
					log.Error("failed to setup manager", "device", m.Device(), "error", err.Error())
					continue
				}
				log.Info("setup device manager", "device", m.Device())
			}

			if viper.GetString("influx.address") != "" {
				log.Info("starting influx exporter", "address", viper.GetString("influx.address"))
				if errC, err := influx.RunExporter(viper.GetString("influx.address"), viper.GetString("influx.token"), viper.GetString("influx.org"), viper.GetString("influx.bucket"), viper.GetInt("influx.interval")); err != nil {
					log.Error("failed to start influx exporter", "error", err.Error())
				} else {
					log.Info("started influx exporter")
					go func() {
						err := <-errC
						if err != nil {
							log.Error("influx exporter error", "error", err.Error())
						}
					}()
				}
			}

			var action string
			options := make([]huh.Option[string], 0)
			for name, _ := range forms.Forms() {
				options = append(options, huh.NewOption(name, name))
			}
			options = append(options, huh.NewOption("Exit", "Exit"))
			for {
				if err := huh.NewSelect[string]().
					Title("Select Action").
					Options(options...).
					Value(&action).
					Run(); err != nil {
					log.Fatal("action selection form was canceled")
				}
				if action == "Exit" {
					break
				}
				if f := forms.Form(action); f != nil {
					f()
				}
			}
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatalln("net4me failed to exectue command")
	}
}
