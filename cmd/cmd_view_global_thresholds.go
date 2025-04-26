package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// runViewGlobalThresholdsCmd displays global thresholds
func runViewGlobalThresholdsCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Load levers configuration
	leversConfig, err := configLoader.LoadLeversConfig()
	if err != nil {
		fmt.Printf("Error loading levers config: %v\n", err)
		os.Exit(1)
	}

	// Display global thresholds
	fmt.Println("Global Thresholds:")
	fmt.Println("-----------------")
	fmt.Printf("Green:  %d-%d\n", leversConfig.Global.Thresholds.Green.Min, leversConfig.Global.Thresholds.Green.Max)
	fmt.Printf("Yellow: %d-%d\n", leversConfig.Global.Thresholds.Yellow.Min, leversConfig.Global.Thresholds.Yellow.Max)
	fmt.Printf("Red:    %d-%d\n", leversConfig.Global.Thresholds.Red.Min, leversConfig.Global.Thresholds.Red.Max)
}
