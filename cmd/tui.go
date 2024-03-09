package cmd

import (
	"fmt"
	"os"

	"github.com/braheezy/chip-8/internal/tui"
	"github.com/charmbracelet/log"

	"github.com/spf13/cobra"
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

		tui.Run(args[0], logger)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
