package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// runViewCategoryThresholdsCmd displays category-specific thresholds
func runViewCategoryThresholdsCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Load levers configuration
	leversConfig, err := configLoader.LoadLeversConfig()
	if err != nil {
		fmt.Printf("Error loading levers config: %v\n", err)
		os.Exit(1)
	}

	// Display category-specific thresholds
	fmt.Println("Category-Specific Thresholds:")
	fmt.Println("----------------------------")
	if len(leversConfig.Weights.CategoryThresholds) == 0 {
		fmt.Println("No category-specific thresholds defined.")
	} else {
		for category, thresholds := range leversConfig.Weights.CategoryThresholds {
			fmt.Printf("%s:\n", category)
			fmt.Printf("  Green:  %d-%d\n", thresholds.Green.Min, thresholds.Green.Max)
			fmt.Printf("  Yellow: %d-%d\n", thresholds.Yellow.Min, thresholds.Yellow.Max)
			fmt.Printf("  Red:    %d-%d\n", thresholds.Red.Min, thresholds.Red.Max)
		}
	}
}
