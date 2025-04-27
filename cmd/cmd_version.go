package main

import (
	"fmt"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// Add version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version and build information of the Pulse CLI.`,
	Run:   runVersionCmd,
}

func runVersionCmd(cmd *cobra.Command, args []string) {
	fmt.Printf("Pulse CLI version %s+%s\n", pulse.Version, pulse.Build)
}
