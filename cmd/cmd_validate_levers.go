package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// runValidateLeversCmd validates both category weights and threshold configurations
func runValidateLeversCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Load levers configuration
	leversConfig, err := configLoader.LoadLeversConfig()
	if err != nil {
		fmt.Printf("Error loading levers config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Running complete validation of levers configuration...")
	fmt.Println()

	// First validate weights
	fmt.Println("=== Category Weights Validation ===")

	// Sum up the category weights
	var totalWeight float64
	for _, weight := range leversConfig.Weights.Categories {
		totalWeight += weight
	}

	// Check if the weights add up to 100% (1.0)
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
	weightsValid := totalWeight >= 1.0-epsilon && totalWeight <= 1.0+epsilon

	if weightsValid {
		fmt.Println("✅ Weights validation PASSED: Category weights add up to 100%")
	} else {
		fmt.Printf("❌ Weights validation FAILED: Category weights add up to %.0f%%, expected 100%%\n", totalWeight*100)
	}

	fmt.Println()
	fmt.Println("=== Threshold Ranges Validation ===")

	// Display the current thresholds
	fmt.Printf("Global Thresholds:\n")
	fmt.Printf("Green:  %d-%d\n", leversConfig.Global.Thresholds.Green.Min, leversConfig.Global.Thresholds.Green.Max)
	fmt.Printf("Yellow: %d-%d\n", leversConfig.Global.Thresholds.Yellow.Min, leversConfig.Global.Thresholds.Yellow.Max)
	fmt.Printf("Red:    %d-%d\n", leversConfig.Global.Thresholds.Red.Min, leversConfig.Global.Thresholds.Red.Max)
	fmt.Println()

	// Validate thresholds
	thresholdsValid := true
	var errors []string

	// Validate global thresholds
	// 1. Check that min <= max for each range
	if leversConfig.Global.Thresholds.Green.Min > leversConfig.Global.Thresholds.Green.Max {
		thresholdsValid = false
		errors = append(errors, fmt.Sprintf("Green threshold min (%d) must be less than or equal to max (%d)",
			leversConfig.Global.Thresholds.Green.Min, leversConfig.Global.Thresholds.Green.Max))
	}

	if leversConfig.Global.Thresholds.Yellow.Min > leversConfig.Global.Thresholds.Yellow.Max {
		thresholdsValid = false
		errors = append(errors, fmt.Sprintf("Yellow threshold min (%d) must be less than or equal to max (%d)",
			leversConfig.Global.Thresholds.Yellow.Min, leversConfig.Global.Thresholds.Yellow.Max))
	}

	if leversConfig.Global.Thresholds.Red.Min > leversConfig.Global.Thresholds.Red.Max {
		thresholdsValid = false
		errors = append(errors, fmt.Sprintf("Red threshold min (%d) must be less than or equal to max (%d)",
			leversConfig.Global.Thresholds.Red.Min, leversConfig.Global.Thresholds.Red.Max))
	}

	// 2. Check that ranges don't overlap
	if leversConfig.Global.Thresholds.Yellow.Max >= leversConfig.Global.Thresholds.Green.Min {
		thresholdsValid = false
		errors = append(errors, fmt.Sprintf("Yellow threshold max (%d) must be less than Green threshold min (%d)",
			leversConfig.Global.Thresholds.Yellow.Max, leversConfig.Global.Thresholds.Green.Min))
	}

	if leversConfig.Global.Thresholds.Red.Max >= leversConfig.Global.Thresholds.Yellow.Min {
		thresholdsValid = false
		errors = append(errors, fmt.Sprintf("Red threshold max (%d) must be less than Yellow threshold min (%d)",
			leversConfig.Global.Thresholds.Red.Max, leversConfig.Global.Thresholds.Yellow.Min))
	}

	// 3. Check that ranges cover the entire range from 0 to 100
	if leversConfig.Global.Thresholds.Red.Min > 0 {
		thresholdsValid = false
		errors = append(errors, fmt.Sprintf("Red threshold min (%d) should be 0 to cover the entire range",
			leversConfig.Global.Thresholds.Red.Min))
	}

	if leversConfig.Global.Thresholds.Green.Max < 100 {
		thresholdsValid = false
		errors = append(errors, fmt.Sprintf("Green threshold max (%d) should be 100 to cover the entire range",
			leversConfig.Global.Thresholds.Green.Max))
	}

	// Also validate category-specific thresholds if they exist
	if len(leversConfig.Weights.CategoryThresholds) > 0 {
		fmt.Println("Category-Specific Thresholds:")

		for category, thresholds := range leversConfig.Weights.CategoryThresholds {
			fmt.Printf("%s:\n", category)
			fmt.Printf("  Green:  %d-%d\n", thresholds.Green.Min, thresholds.Green.Max)
			fmt.Printf("  Yellow: %d-%d\n", thresholds.Yellow.Min, thresholds.Yellow.Max)
			fmt.Printf("  Red:    %d-%d\n", thresholds.Red.Min, thresholds.Red.Max)

			// 1. Check that min <= max for each range
			if thresholds.Green.Min > thresholds.Green.Max {
				thresholdsValid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Green threshold min (%d) must be less than or equal to max (%d)",
					category, thresholds.Green.Min, thresholds.Green.Max))
			}

			if thresholds.Yellow.Min > thresholds.Yellow.Max {
				thresholdsValid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Yellow threshold min (%d) must be less than or equal to max (%d)",
					category, thresholds.Yellow.Min, thresholds.Yellow.Max))
			}

			if thresholds.Red.Min > thresholds.Red.Max {
				thresholdsValid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Red threshold min (%d) must be less than or equal to max (%d)",
					category, thresholds.Red.Min, thresholds.Red.Max))
			}

			// 2. Check that ranges don't overlap
			if thresholds.Yellow.Max >= thresholds.Green.Min {
				thresholdsValid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Yellow threshold max (%d) must be less than Green threshold min (%d)",
					category, thresholds.Yellow.Max, thresholds.Green.Min))
			}

			if thresholds.Red.Max >= thresholds.Yellow.Min {
				thresholdsValid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Red threshold max (%d) must be less than Yellow threshold min (%d)",
					category, thresholds.Red.Max, thresholds.Yellow.Min))
			}

			// 3. Check that ranges cover the entire range from 0 to 100
			if thresholds.Red.Min > 0 {
				thresholdsValid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Red threshold min (%d) should be 0 to cover the entire range",
					category, thresholds.Red.Min))
			}

			if thresholds.Green.Max < 100 {
				thresholdsValid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Green threshold max (%d) should be 100 to cover the entire range",
					category, thresholds.Green.Max))
			}
		}
	}

	// Display threshold validation result
	if thresholdsValid {
		fmt.Println("✅ Thresholds validation PASSED: All threshold ranges are valid and don't overlap")
	} else {
		fmt.Println("❌ Thresholds validation FAILED:")
		for _, err := range errors {
			fmt.Printf("   - %s\n", err)
		}
	}

	// Overall validation result
	fmt.Println()
	fmt.Println("=== Overall Validation Result ===")
	if weightsValid && thresholdsValid {
		fmt.Println("✅ All validations PASSED")
	} else {
		fmt.Println("❌ Some validations FAILED")
		os.Exit(1)
	}
}
