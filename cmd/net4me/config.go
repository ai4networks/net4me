package main

import (
	"strings"

	"github.com/spf13/viper"
)

func init() {
	viper.AddConfigPath("/etc/net4me")
	viper.AddConfigPath("$HOME/.net4me")
	viper.AddConfigPath(".")
	viper.SetConfigName("net4me")
	viper.SetConfigType("toml")

	rootCmd.PersistentFlags().StringP("config", "c", "", "path to net4me configuration file (default auto)")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose logging")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()
}
