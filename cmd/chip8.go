package cmd

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/braheezy/chip-8/internal/chip8"

	"github.com/charmbracelet/log"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "chip8",
	Short: "chip8 is a useful CHIP-8 interpreter",
	Long:  `Run CHIP-8 programs from the CLI quickly`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(args[0])
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var debug bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Show debug messages")

	cobra.OnInitialize(initConfig)
}

func run(romFilePath string) {
	chipFileName := filepath.Base(romFilePath)
	chipData, err := os.ReadFile(romFilePath)
	if err != nil {
		log.Fatal(err)
	}

	app := chip8.NewDefaultApp(&chipData)
	if debug {
		app.Chip8.Logger.SetLevel(log.DebugLevel)
	}
	app.Chip8.Options.DisplayScaleFactor = viper.GetInt("displayScaleFactor")

	ebiten.SetWindowSize(chip8.DisplayWidth*app.Chip8.Options.DisplayScaleFactor, chip8.DisplayHeight*app.Chip8.Options.DisplayScaleFactor)
	ebiten.SetWindowTitle(chipFileName)
	ebiten.SetTPS(ebiten.SyncWithFPS)

	if err := ebiten.RunGame(app); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}

}
