package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

func runCreateMetricFileCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Create the metric file
	fileName := args[0]
	err := configLoader.CreateMetricFile(fileName)
	if err != nil {
		fmt.Printf("Error creating metric file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Metric file '%s' created successfully.\n", fileName)
}
