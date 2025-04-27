package main

import (
	"github.com/spf13/cobra"
)

// Add list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available metrics or categories",
	Long:  `List all available metrics or categories.`,
}
