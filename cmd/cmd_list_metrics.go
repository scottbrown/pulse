package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

func runListMetricsCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Load configurations
	metricsConfig, err := configLoader.LoadMetricsConfig()
	if err != nil {
		fmt.Printf("Error loading metrics config: %v\n", err)
		os.Exit(1)
	}

	leversConfig, err := configLoader.LoadLeversConfig()
	if err != nil {
		fmt.Printf("Error loading levers config: %v\n", err)
		os.Exit(1)
	}

	metricsData, err := configLoader.LoadMetricsData()
	if err != nil {
		fmt.Printf("Error loading metrics data: %v\n", err)
		os.Exit(1)
	}

	// Initialize the metrics processor
	metricsProcessor := pulse.NewMetricsProcessor(metricsConfig, leversConfig, metricsData)

	// Get all metrics
	metrics := metricsProcessor.GetAllMetrics()

	// Display metrics
	fmt.Println("Available Metrics:")
	fmt.Println("------------------")
	for _, metric := range metrics {
		fmt.Printf("%s: %.2f (as of %s)\n", metric.Reference, metric.Value, metric.Timestamp.Format("2006-01-02"))
	}
}
