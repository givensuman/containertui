package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/config"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui"
	"github.com/spf13/cobra"
)

func resolveStartupTab(configTab, subcommandTab string) string {
	startupTab := strings.TrimSpace(configTab)

	if strings.TrimSpace(subcommandTab) != "" {
		startupTab = strings.TrimSpace(subcommandTab)
	}

	return startupTab
}

func runContainertui(cmd *cobra.Command, tabName string, noNerdFonts bool, configPath string, colorsFlag []string, jsonFormat bool) error {
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

	if jsonFormat {
		cfg.InspectionFormat = "json"
	}

	cfg.StartupTab = resolveStartupTab(cfg.StartupTab, tabName)

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

	state.SetConfig(cfg)

	// Initialize the shared Docker client
	if err := state.InitializeClient(); err != nil {
		return fmt.Errorf("failed to initialize Docker client: %w", err)
	}
	defer func() {
		if err := state.CloseClient(); err != nil {
			log.Printf("error closing Docker client: %v", err)
		}
	}()

	state.InitializeLog()

	// Start the UI
	if err := ui.Start(); err != nil {
		return fmt.Errorf("failed to run application: %w", err)
	}

	return nil
}

func main() {
	var noNerdFonts bool
	var configPath string
	var colorsFlag []string
	var jsonFormat bool

	// Create subcommand runner factory
	makeSubcommand := func(tabName string, use string, short string) *cobra.Command {
		return &cobra.Command{
			Use:   use,
			Short: short,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runContainertui(cmd, tabName, noNerdFonts, configPath, colorsFlag, jsonFormat)
			},
		}
	}

	rootCmd := &cobra.Command{
		Use:   "containertui",
		Short: "a tui for managing container lifecycles",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContainertui(cmd, "", noNerdFonts, configPath, colorsFlag, jsonFormat)
		},
	}

	// Add subcommands
	rootCmd.AddCommand(makeSubcommand("containers", "containers", "launch containertui to the containers tab"))
	rootCmd.AddCommand(makeSubcommand("images", "images", "launch containertui to the images tab"))
	rootCmd.AddCommand(makeSubcommand("volumes", "volumes", "launch containertui to the volumes tab"))
	rootCmd.AddCommand(makeSubcommand("networks", "networks", "launch containertui to the networks tab"))
	rootCmd.AddCommand(makeSubcommand("browse", "browse", "launch containertui to the browse tab"))

	// Add global flags to root command
	rootCmd.PersistentFlags().BoolVar(&noNerdFonts, "no-nerd-fonts", false, "disable nerd fonts")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "path to config file")
	rootCmd.PersistentFlags().StringSliceVar(&colorsFlag, "colors", nil, "color overrides (format: --colors 'primary=#b4befe' --colors 'warning=#f9e2af,success=#a6e3a1')")
	rootCmd.PersistentFlags().BoolVar(&jsonFormat, "json", false, "use JSON format for inspection output")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
