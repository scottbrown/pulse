package main

import (
	"github.com/spf13/cobra"
)

// Add metrics subcommand
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Manage metrics and metric files",
	Long:  `Commands for managing metrics and metric files.`,
}
