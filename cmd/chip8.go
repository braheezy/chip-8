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
	Use:   "chip8 romFile",
	Short: "chip8 is a useful CHIP-8 interpreter",
	Long:  `Run CHIP-8 programs from the CLI with ease`,
	Args: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("list-modes") {
			return nil
		}
		if viper.GetBool("write-config") {
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("requires ROM file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger := newDefaultLogger()
		if viper.GetBool("write-config") {
			// Don't get caught in forever loop of writing config file
			viper.Set("write-config", false)
			err := viper.SafeWriteConfig()
			if _, ok := err.(viper.ConfigFileAlreadyExistsError); ok {
				logger.Info("Config file already exists, overwriting...")
				err = viper.WriteConfig()
				if err != nil {
					logger.Fatal(err)
				}
			}
		} else if viper.GetBool("list-modes") {
			fmt.Println("Supported CHIP-8 variants:")
			for _, mode := range interpreter.SupportedModes {
				fmt.Println(mode)
			}
		} else {
			run(args[0], logger)
		}
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var debug bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Show debug messages")

	rootCmd.Flags().BoolP("cosmac", "c", false, "Run in COSMAC VIP mode")
	viper.BindPFlag("cosmac-vip.enabled", rootCmd.Flags().Lookup("cosmac"))

	rootCmd.Flags().Bool("write-config", false, "Write current config to default location. Existing config file will be overwritten!")
	viper.BindPFlag("write-config", rootCmd.Flags().Lookup("write-config"))

	rootCmd.Flags().Bool("list-modes", false, "Show supported CHIP-8 variants")
	viper.BindPFlag("list-modes", rootCmd.Flags().Lookup("list-modes"))

	cobra.OnInitialize(initConfig)
}

func run(romFilePath string, logger *log.Logger) {
	chipFileName := filepath.Base(romFilePath)
	chipData, err := os.ReadFile(romFilePath)
	if err != nil {
		logger.Fatal(err)
	}
	if debug {
		logger.SetLevel(log.DebugLevel)
	}

	app := interpreter.NewCHIP8(&chipData)
	app.Logger = logger
	viper.Unmarshal(&app.Options)

	if viper.GetBool("cosmac-vip.enabled") {
		logger.Info("COSMAC VIP mode enabled")
		app.Options.CosmacQuirks.EnableAll()
	}

	ebiten.SetWindowSize(interpreter.DisplayWidth*app.Options.DisplayScaleFactor, interpreter.DisplayHeight*app.Options.DisplayScaleFactor)
	ebiten.SetWindowTitle(chipFileName)
	ebiten.SetTPS(ebiten.SyncWithFPS)

	if err := ebiten.RunGame(app); err != nil && err != ebiten.Termination {
		logger.Fatal(err)
	}

}
