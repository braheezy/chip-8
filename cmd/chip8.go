package cmd

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/braheezy/chip-8/internal/interpreter"

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

	app := interpreter.NewCHIP8(&chipData)
	logger := newDefaultLogger()
	if debug {
		logger.SetLevel(log.DebugLevel)
	}
	app.Logger = logger
	viper.Unmarshal(&app.Options)

	if viper.IsSet("cosmac-vip") {
		logger.Info("Cosmac VIP mode enabled")
	}

	ebiten.SetWindowSize(interpreter.DisplayWidth*app.Options.DisplayScaleFactor, interpreter.DisplayHeight*app.Options.DisplayScaleFactor)
	ebiten.SetWindowTitle(chipFileName)
	ebiten.SetTPS(ebiten.SyncWithFPS)

	if err := ebiten.RunGame(app); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}

}
