package main

import (
	"fmt"
	"log"

	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/config"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui"
	"github.com/spf13/cobra"
)

func main() {
	var noNerdFonts bool
	var configPath string
	var colorsFlag []string

	rootCmd := &cobra.Command{
		Use:   "containertui",
		Short: "a tui for managing container lifecycles",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg *config.Config
			var err error
			if configPath != "" {
				cfg, err = config.LoadFromFile(configPath)
				if err != nil {
					return err
				}
			} else {
				cfg = config.DefaultConfig()
			}

			if noNerdFonts {
				cfg.NoNerdFonts = true
			}

			if len(colorsFlag) > 0 {
				colorOverrides, err := colors.ParseColors(colorsFlag)
				if err != nil {
					return fmt.Errorf("failed to parse colors: %w", err)
				}

				if colorOverrides.Primary.IsAssigned() {
					cfg.Colors.Primary = colorOverrides.Primary
				}
				if colorOverrides.Yellow.IsAssigned() {
					cfg.Colors.Yellow = colorOverrides.Yellow
				}
				if colorOverrides.Green.IsAssigned() {
					cfg.Colors.Green = colorOverrides.Green
				}
				if colorOverrides.Blue.IsAssigned() {
					cfg.Colors.Blue = colorOverrides.Blue
				}
				if colorOverrides.White.IsAssigned() {
					cfg.Colors.White = colorOverrides.White
				}
			}

			context.SetConfig(cfg)

			context.InitializeClient()
			defer context.CloseClient()

			if err := ui.Start(); err != nil {
				return fmt.Errorf("failed to run application: %w", err)
			}

			return nil
		},
	}

	rootCmd.Flags().BoolVar(&noNerdFonts, "no-nerd-fonts", false, "disable nerd fonts")
	rootCmd.Flags().StringVar(&configPath, "config", "", "path to config file")
	rootCmd.Flags().StringSliceVar(&colorsFlag, "colors", nil, "color overrides (format: --colors 'primary=#b4befe' --colors 'yellow=#f9e2af,green=#a6e3a1')")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
