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
					cfg.Theme.Primary = colorOverrides.Primary
				}
				if colorOverrides.Border.IsAssigned() {
					cfg.Theme.Border = colorOverrides.Border
				}
				if colorOverrides.Text.IsAssigned() {
					cfg.Theme.Text = colorOverrides.Text
				}
				if colorOverrides.Muted.IsAssigned() {
					cfg.Theme.Muted = colorOverrides.Muted
				}
				if colorOverrides.Selected.IsAssigned() {
					cfg.Theme.Selected = colorOverrides.Selected
				}
				if colorOverrides.Success.IsAssigned() {
					cfg.Theme.Success = colorOverrides.Success
				}
				if colorOverrides.Warning.IsAssigned() {
					cfg.Theme.Warning = colorOverrides.Warning
				}
				if colorOverrides.Error.IsAssigned() {
					cfg.Theme.Error = colorOverrides.Error
				}
			}

			context.SetConfig(cfg)

			// Initialize the shared Docker client
			if err := context.InitializeClient(); err != nil {
				return fmt.Errorf("failed to initialize Docker client: %w", err)
			}
			defer func() {
				if err := context.CloseClient(); err != nil {
					log.Printf("error closing Docker client: %v", err)
				}
			}()

			context.InitializeLog()

			// Start the UI
			if err := ui.Start(); err != nil {
				return fmt.Errorf("failed to run application: %w", err)
			}

			return nil
		},
	}

	rootCmd.Flags().BoolVar(&noNerdFonts, "no-nerd-fonts", false, "disable nerd fonts")
	rootCmd.Flags().StringVar(&configPath, "config", "", "path to config file")
	rootCmd.Flags().StringSliceVar(&colorsFlag, "colors", nil, "color overrides (format: --colors 'primary=#b4befe' --colors 'warning=#f9e2af,success=#a6e3a1')")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
