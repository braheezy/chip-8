package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/braheezy/chip-8/internal/interpreter"
	"github.com/charmbracelet/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var tuiCmd = &cobra.Command{
	Use:   "tui <rom>",
	Short: "Run in TUI mode",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return fmt.Errorf("requires ROM file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger := newDefaultLogger()
		logFile, err := os.Create("chip8.log")
		if err != nil {
			logger.Fatalf("error opening file for logging: %v", err)
		}
		defer logFile.Close()
		if debug {
			logger.SetLevel(log.DebugLevel)
		}
		logger.SetOutput(logFile)

		chipFileName := filepath.Base(args[0])
		chipData, err := os.ReadFile(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		opts := interpreter.DefaultCHIP8Options()
		viper.Unmarshal(&opts)
		chip8 := interpreter.NewCHIP8(&chipData, opts)
		chip8.Logger = logger

		if viper.GetBool("cosmac-vip.enabled") {
			logger.Info("COSMAC VIP mode enabled")
			chip8.Options.CosmacQuirks.EnableAll()
		}

		interpreter.RunTUI(chip8, chipFileName)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
