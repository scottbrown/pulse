package main

import (
	"github.com/spf13/cobra"
)

// Add levers command
var leversCmd = &cobra.Command{
	Use:   "levers",
	Short: "View configuration levers",
	Long:  `View configuration levers that affect scoring and reporting.`,
}
