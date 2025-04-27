package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

var categoryWeightsCmd = &cobra.Command{
	Use:   "category-weights",
	Short: "View category weights",
	Long:  `View weights assigned to each category for overall score calculation.`,
	Run:   runViewCategoryWeightsCmd,
}

// runViewCategoryWeightsCmd displays category weights
func runViewCategoryWeightsCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Load levers configuration
	leversConfig, err := configLoader.LoadLeversConfig()
	if err != nil {
		fmt.Printf("Error loading levers config: %v\n", err)
		os.Exit(1)
	}

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
}
