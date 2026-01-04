package main

import (
	"log"

	"github.com/givensuman/containertui/internal/config"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui"
	"github.com/spf13/cobra"
)

func main() {
	var noNerdFonts bool
	var configPath string

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
			}

			if noNerdFonts {
				cfg.NoNerdFonts = true
			}

			context.SetConfig(cfg)

			context.InitializeClient()
			defer context.CloseClient()

			ui.Start()
			return nil
		},
	}

	rootCmd.Flags().BoolVar(&noNerdFonts, "no-nerd-fonts", false, "Disable nerd fonts")
	rootCmd.Flags().StringVar(&configPath, "config", "", "Path to config file")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
