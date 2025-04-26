package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

func runUpdateCmd(cmd *cobra.Command, args []string) {
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

	// Validate metric reference format
	if !strings.Contains(metricRef, ".") || len(strings.Split(metricRef, ".")) != 3 {
		fmt.Printf("Error: Invalid metric reference format. Expected format: category.TYPE.name\n")
		os.Exit(1)
	}

	// Parse and validate the metric value
	value, err := strconv.ParseFloat(metricVal, 64)
	if err != nil {
		fmt.Printf("Error parsing metric value: %v\n", err)
		os.Exit(1)
	}

	// Check for reasonable bounds on the value
	if math.IsNaN(value) || math.IsInf(value, 0) || value < -1000000 || value > 1000000 {
		fmt.Printf("Error: Metric value out of reasonable bounds\n")
		os.Exit(1)
	}

	// Update the metric
	err = metricsProcessor.UpdateMetric(metricRef, value)
	if err != nil {
		fmt.Printf("Error updating metric: %v\n", err)
		os.Exit(1)
	}

	// Save the updated metrics data
	err = configLoader.SaveMetricsData(metricsData)
	if err != nil {
		fmt.Printf("Error saving metrics data: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Metric %s updated to %s\n", metricRef, metricVal)
}
