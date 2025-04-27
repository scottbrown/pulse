package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

var validateThresholdsCmd = &cobra.Command{
	Use:   "validate-thresholds",
	Short: "Validate threshold configurations",
	Long:  `Validate that global and category-specific thresholds do not overlap and follow the correct order (Red < Yellow < Green).`,
	Run:   runValidateThresholdsCmd,
}

// runValidateThresholdsCmd validates that threshold ranges are valid and don't overlap
func runValidateThresholdsCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Load levers configuration
	leversConfig, err := configLoader.LoadLeversConfig()
	if err != nil {
		fmt.Printf("Error loading levers config: %v\n", err)
		os.Exit(1)
	}

	// Display the current thresholds
	fmt.Println("Global Thresholds Validation:")
	fmt.Println("--------------------------")
	fmt.Printf("Green:  %d-%d\n", leversConfig.Global.Thresholds.Green.Min, leversConfig.Global.Thresholds.Green.Max)
	fmt.Printf("Yellow: %d-%d\n", leversConfig.Global.Thresholds.Yellow.Min, leversConfig.Global.Thresholds.Yellow.Max)
	fmt.Printf("Red:    %d-%d\n", leversConfig.Global.Thresholds.Red.Min, leversConfig.Global.Thresholds.Red.Max)
	fmt.Println()

	// Validate thresholds
	valid := true
	var errors []string

	// Validate global thresholds
	// 1. Check that min <= max for each range
	if leversConfig.Global.Thresholds.Green.Min > leversConfig.Global.Thresholds.Green.Max {
		valid = false
		errors = append(errors, fmt.Sprintf("Green threshold min (%d) must be less than or equal to max (%d)",
			leversConfig.Global.Thresholds.Green.Min, leversConfig.Global.Thresholds.Green.Max))
	}

	if leversConfig.Global.Thresholds.Yellow.Min > leversConfig.Global.Thresholds.Yellow.Max {
		valid = false
		errors = append(errors, fmt.Sprintf("Yellow threshold min (%d) must be less than or equal to max (%d)",
			leversConfig.Global.Thresholds.Yellow.Min, leversConfig.Global.Thresholds.Yellow.Max))
	}

	if leversConfig.Global.Thresholds.Red.Min > leversConfig.Global.Thresholds.Red.Max {
		valid = false
		errors = append(errors, fmt.Sprintf("Red threshold min (%d) must be less than or equal to max (%d)",
			leversConfig.Global.Thresholds.Red.Min, leversConfig.Global.Thresholds.Red.Max))
	}

	// 2. Check that ranges don't overlap
	if leversConfig.Global.Thresholds.Yellow.Max >= leversConfig.Global.Thresholds.Green.Min {
		valid = false
		errors = append(errors, fmt.Sprintf("Yellow threshold max (%d) must be less than Green threshold min (%d)",
			leversConfig.Global.Thresholds.Yellow.Max, leversConfig.Global.Thresholds.Green.Min))
	}

	if leversConfig.Global.Thresholds.Red.Max >= leversConfig.Global.Thresholds.Yellow.Min {
		valid = false
		errors = append(errors, fmt.Sprintf("Red threshold max (%d) must be less than Yellow threshold min (%d)",
			leversConfig.Global.Thresholds.Red.Max, leversConfig.Global.Thresholds.Yellow.Min))
	}

	// 3. Check that ranges cover the entire range from 0 to 100
	if leversConfig.Global.Thresholds.Red.Min > 0 {
		valid = false
		errors = append(errors, fmt.Sprintf("Red threshold min (%d) should be 0 to cover the entire range",
			leversConfig.Global.Thresholds.Red.Min))
	}

	if leversConfig.Global.Thresholds.Green.Max < 100 {
		valid = false
		errors = append(errors, fmt.Sprintf("Green threshold max (%d) should be 100 to cover the entire range",
			leversConfig.Global.Thresholds.Green.Max))
	}

	// Also validate category-specific thresholds if they exist
	if len(leversConfig.Weights.CategoryThresholds) > 0 {
		fmt.Println("Category-Specific Thresholds Validation:")
		fmt.Println("-------------------------------------")

		for category, thresholds := range leversConfig.Weights.CategoryThresholds {
			fmt.Printf("%s:\n", category)
			fmt.Printf("  Green:  %d-%d\n", thresholds.Green.Min, thresholds.Green.Max)
			fmt.Printf("  Yellow: %d-%d\n", thresholds.Yellow.Min, thresholds.Yellow.Max)
			fmt.Printf("  Red:    %d-%d\n", thresholds.Red.Min, thresholds.Red.Max)

			// 1. Check that min <= max for each range
			if thresholds.Green.Min > thresholds.Green.Max {
				valid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Green threshold min (%d) must be less than or equal to max (%d)",
					category, thresholds.Green.Min, thresholds.Green.Max))
			}

			if thresholds.Yellow.Min > thresholds.Yellow.Max {
				valid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Yellow threshold min (%d) must be less than or equal to max (%d)",
					category, thresholds.Yellow.Min, thresholds.Yellow.Max))
			}

			if thresholds.Red.Min > thresholds.Red.Max {
				valid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Red threshold min (%d) must be less than or equal to max (%d)",
					category, thresholds.Red.Min, thresholds.Red.Max))
			}

			// 2. Check that ranges don't overlap
			if thresholds.Yellow.Max >= thresholds.Green.Min {
				valid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Yellow threshold max (%d) must be less than Green threshold min (%d)",
					category, thresholds.Yellow.Max, thresholds.Green.Min))
			}

			if thresholds.Red.Max >= thresholds.Yellow.Min {
				valid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Red threshold max (%d) must be less than Yellow threshold min (%d)",
					category, thresholds.Red.Max, thresholds.Yellow.Min))
			}

			// 3. Check that ranges cover the entire range from 0 to 100
			if thresholds.Red.Min > 0 {
				valid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Red threshold min (%d) should be 0 to cover the entire range",
					category, thresholds.Red.Min))
			}

			if thresholds.Green.Max < 100 {
				valid = false
				errors = append(errors, fmt.Sprintf("Category '%s': Green threshold max (%d) should be 100 to cover the entire range",
					category, thresholds.Green.Max))
			}
		}
	}

	fmt.Println()

	// Display validation result
	if valid {
		fmt.Println("✅ Validation PASSED: All threshold ranges are valid and don't overlap")
	} else {
		fmt.Println("❌ Validation FAILED:")
		for _, err := range errors {
			fmt.Printf("   - %s\n", err)
		}
		fmt.Println()
		fmt.Println("Threshold ranges should follow these rules:")
		fmt.Println("  1. Min must be less than or equal to Max for each range")
		fmt.Println("  2. Ranges must not overlap (Red.Max < Yellow.Min, Yellow.Max < Green.Min)")
		fmt.Println("  3. Ranges should cover the entire range from 0 to 100")
		fmt.Println()
		fmt.Println("Example of valid threshold ranges:")
		fmt.Println("  Green:  80-100")
		fmt.Println("  Yellow: 60-79")
		fmt.Println("  Red:    0-59")
		os.Exit(1)
	}
}
