package cmd

import (
	"fmt"

	"github.com/braheezy/chip-8/internal/interpreter"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list-modes",
	Short: "Show supported CHIP-8 variants",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Supported CHIP-8 variants:")
		for _, mode := range interpreter.SupportedModes {
			fmt.Println(mode)
		}
	},
}
