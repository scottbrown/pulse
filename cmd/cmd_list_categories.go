package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// Add categories subcommand
var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List all available categories",
	Long:  `List all available categories with their KPIs and KRIs.`,
	Run:   runListCategoriesCmd,
}

func runListCategoriesCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Load configurations
	metricsConfig, err := configLoader.LoadMetricsConfig()
	if err != nil {
		fmt.Printf("Error loading metrics config: %v\n", err)
		os.Exit(1)
	}

	// Initialize the metrics processor (without loading other configs since we don't need them)
	metricsProcessor := pulse.NewMetricsProcessor(metricsConfig, nil, nil)

	// Get all categories
	categories := metricsProcessor.GetAllCategories()

	// Display categories
	fmt.Println("Available Categories:")
	fmt.Println("--------------------")
	for _, category := range categories {
		fmt.Printf("%s (%s): %s\n", category.Name, category.ID, category.Description)

		fmt.Println("  KPIs:")
		for _, kpi := range category.KPIs {
			fmt.Printf("  - %s (%s): %s [Target: %.2f %s]\n", kpi.Name, kpi.ID, kpi.Description, kpi.Target, kpi.Unit)
		}

		fmt.Println("  KRIs:")
		for _, kri := range category.KRIs {
			fmt.Printf("  - %s (%s): %s [Threshold: %.2f %s]\n", kri.Name, kri.ID, kri.Description, kri.Threshold, kri.Unit)
		}

		fmt.Println()
	}
}
