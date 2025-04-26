package main

import (
	"fmt"
	"os"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

func runReportCmd(cmd *cobra.Command, args []string) {
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

	// Initialize the score calculator with the specified scoring method
	var scoringMethodEnum pulse.ScoringMethod
	if scoringMethod == "average" {
		scoringMethodEnum = pulse.AverageScoring
	} else {
		// Default to median scoring
		scoringMethodEnum = pulse.MedianScoring
	}

	scoreCalculator := pulse.NewScoreCalculator(metricsProcessor, scoringMethodEnum)

	// Initialize the report generator
	reportGenerator := pulse.NewReportGenerator(scoreCalculator)

	// Generate the report
	var reportContent string
	var reportErr error

	reportFormat := pulse.TextFormat
	if format == "json" {
		reportFormat = pulse.JSONFormat
	}

	if category != "" {
		reportContent, reportErr = reportGenerator.GenerateCategoryReport(category, reportFormat)
	} else {
		reportContent, reportErr = reportGenerator.GenerateOverallReport(reportFormat)
	}

	if reportErr != nil {
		fmt.Printf("Error generating report: %v\n", reportErr)
		os.Exit(1)
	}

	// Output the report
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(reportContent), 0600)
		if err != nil {
			fmt.Printf("Error writing report to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Report written to %s\n", outputFile)
	} else {
		fmt.Println(reportContent)
	}
}
