package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// Add levers subcommands
var allLeversCmd = &cobra.Command{
	Use:   "all",
	Short: "View all configuration levers",
	Long:  `View all configuration levers including global weights, category weights, global thresholds, and scoring bands.`,
	Run:   runViewAllLeversCmd,
}

// runViewAllLeversCmd displays all configuration levers
func runViewAllLeversCmd(cmd *cobra.Command, args []string) {
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
	fmt.Println()

	// Note: Global scoring bands have been removed in favor of per-metric scoring bands
	fmt.Println()

	// Display category weights
	fmt.Println("Category Weights:")
	fmt.Println("----------------")
	if len(leversConfig.Weights.Categories) == 0 {
		fmt.Println("No category weights defined.")
	} else {
		for category, weight := range leversConfig.Weights.Categories {
			fmt.Printf("%s: %.2f\n", category, weight)
		}
	}
	fmt.Println()

	// Display category-specific thresholds
	fmt.Println("Category-Specific Thresholds:")
	fmt.Println("----------------------------")
	if len(leversConfig.Weights.CategoryThresholds) == 0 {
		fmt.Println("No category-specific thresholds defined.")
	} else {
		for category, thresholds := range leversConfig.Weights.CategoryThresholds {
			fmt.Printf("%s:\n", category)
			fmt.Printf("  Green:  %d\n", thresholds.Green)
			fmt.Printf("  Yellow: %d\n", thresholds.Yellow)
			fmt.Printf("  Red:    %d\n", thresholds.Red)
		}
	}
}
