package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// Add report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a security posture report",
	Long:  `Generate a report on the security posture based on the configured KPIs and KRIs.`,
	Run:   runReportCmd,
}

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

	// Check if we have any categories defined
	if len(metricsConfig.Categories) == 0 {
		fmt.Println("No categories defined in metrics configuration.")
		fmt.Println("Please create a metrics.yaml file in your config directory or run 'pulse init' to create default configuration files.")
		os.Exit(0)
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

	// Determine threshold label type
	var thresholdLabelType pulse.ThresholdLabelType
	if strings.EqualFold(thresholdLabels, "text") {
		thresholdLabelType = pulse.TextLabels
	} else {
		// Default to emoji labels
		thresholdLabelType = pulse.EmojiLabels
	}

	// Initialize the report generator
	reportGenerator := pulse.NewReportGenerator(scoreCalculator, thresholdLabelType)

	// Generate the report
	var reportContent string
	var reportErr error

	reportFormat := pulse.TextFormat
	switch format {
	case "json":
		reportFormat = pulse.JSONFormat
	case "table":
		reportFormat = pulse.TableFormat
	}

	if category != "" {
		reportContent, reportErr = reportGenerator.GenerateCategoryReport(category, reportFormat)
	} else {
		reportContent, reportErr = reportGenerator.GenerateOverallReport(reportFormat)
	}

	if reportErr != nil {
		if reportErr.Error() == "no categories found" {
			fmt.Println("No categories defined in metrics configuration.")
			fmt.Println("Please create a metrics.yaml file in your config directory or run 'pulse init' to create default configuration files.")
			os.Exit(0)
		}
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
