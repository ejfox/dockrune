package main

import (
	"fmt"
	"os"

	"github.com/ejfox/dockrune/internal/cmd"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:     "dockrune",
		Short:   "Self-hosted deployment daemon for GitHub webhooks",
		Long:    `dockrune receives GitHub webhooks, detects project types, and runs appropriate deploy commands on your VPS.`,
		Version: version,
	}

	// Add subcommands
	rootCmd.AddCommand(cmd.ServeCmd())
	rootCmd.AddCommand(cmd.InitCmd())
	rootCmd.AddCommand(cmd.DeployCmd())
	rootCmd.AddCommand(cmd.StatusCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
