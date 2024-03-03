package cmd

import (
	"fmt"

	"github.com/spf13/viper"
)

func initConfig() {
	// Add config file search paths
	viper.AddConfigPath("$XDG_CONFIG_HOME/chip8")
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	// Use env vars
	viper.AutomaticEnv()
	viper.SetEnvPrefix("chip8")
	viper.BindEnv("display_scale_factor", "DISPLAY_SCALE_FACTOR")
	viper.BindEnv("throttle_speed", "THROTTLE_SPEED")
	viper.BindEnv("instruction_limit", "INSTRUCTION_LIMIT")

	// Set sane defaults
	viper.SetDefault("display_scale_factor", 10)
}
