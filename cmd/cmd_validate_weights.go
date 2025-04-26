package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// runValidateWeightsCmd validates that category weights add up to 100%
func runValidateWeightsCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Load levers configuration
	leversConfig, err := configLoader.LoadLeversConfig()
	if err != nil {
		fmt.Printf("Error loading levers config: %v\n", err)
		os.Exit(1)
	}

	// Sum up the category weights
	var totalWeight float64
	for _, weight := range leversConfig.Weights.Categories {
		totalWeight += weight
	}

	// Check if the weights add up to 100% (1.0)
	fmt.Println("Category Weights Validation:")
	fmt.Println("--------------------------")

	if len(leversConfig.Weights.Categories) == 0 {
		fmt.Println("No category weights defined.")
		os.Exit(1)
	}

	// Display all category weights
	for category, weight := range leversConfig.Weights.Categories {
		fmt.Printf("%s: %.2f (%.0f%%)\n", category, weight, weight*100)
	}
	fmt.Println()

	// Display the total and validation result
	fmt.Printf("Total weight: %.2f (%.0f%%)\n", totalWeight, totalWeight*100)

	// Use a small epsilon for floating point comparison
	const epsilon = 0.0001
	if totalWeight >= 1.0-epsilon && totalWeight <= 1.0+epsilon {
		fmt.Println("✅ Validation PASSED: Category weights add up to 100%")
	} else {
		fmt.Printf("❌ Validation FAILED: Category weights add up to %.0f%%, expected 100%%\n", totalWeight*100)
		os.Exit(1)
	}
}
