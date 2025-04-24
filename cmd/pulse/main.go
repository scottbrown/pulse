package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// Version information set by build flags
var (
	version = "main"
	build   = "unknown"
)

var (
	configDir     string
	dataDir       string
	category      string
	format        string
	outputFile    string
	metricRef     string
	metricVal     string
	scoringMethod string
)

func main() {
	// Define root command
	rootCmd := &cobra.Command{
		Use:     "pulse",
		Short:   "Pulse - Risk and Performance measurement framework CLI",
		Long:    `A CLI application for reporting on Key Performance Indicators (KPIs) and Key Risk Indicators (KRIs) for security programs.`,
		Version: fmt.Sprintf("%s (%s)", version, build),
	}

	// Set up default directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}

	defaultConfigDir := filepath.Join(homeDir, ".pulse", "config")
	defaultDataDir := filepath.Join(homeDir, ".pulse", "data")

	// Add persistent flags for config and data directories
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", defaultConfigDir, "Directory containing configuration files")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", defaultDataDir, "Directory containing data files")

	// Add report command
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a security posture report",
		Long:  `Generate a report on the security posture based on the configured KPIs and KRIs.`,
		Run:   runReportCmd,
	}

	reportCmd.Flags().StringVarP(&category, "category", "c", "", "Generate report for a specific category")
	reportCmd.Flags().StringVarP(&format, "format", "f", "text", "Report format (text or json)")
	reportCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	reportCmd.Flags().StringVar(&scoringMethod, "scoring-method", "median", "Scoring method to use (median or average)")

	// Add update command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a metric value",
		Long:  `Update the value of a specific metric.`,
		Run:   runUpdateCmd,
	}

	updateCmd.Flags().StringVarP(&metricRef, "metric", "m", "", "Metric reference (e.g., app_sec.KPI.vuln_remediation_time)")
	updateCmd.Flags().StringVarP(&metricVal, "value", "v", "", "Metric value")
	updateCmd.MarkFlagRequired("metric")
	updateCmd.MarkFlagRequired("value")

	// Add list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available metrics or categories",
		Long:  `List all available metrics or categories.`,
	}

	// Add levers command
	leversCmd := &cobra.Command{
		Use:   "levers",
		Short: "View configuration levers",
		Long:  `View configuration levers that affect scoring and reporting.`,
	}

	// Add levers subcommands
	allLeversCmd := &cobra.Command{
		Use:   "all",
		Short: "View all configuration levers",
		Long:  `View all configuration levers including global weights, category weights, global thresholds, and scoring bands.`,
		Run:   runViewAllLeversCmd,
	}

	globalThresholdsCmd := &cobra.Command{
		Use:   "global-thresholds",
		Short: "View global thresholds",
		Long:  `View global thresholds for the traffic light model (green, yellow, red).`,
		Run:   runViewGlobalThresholdsCmd,
	}

	scoringBandsCmd := &cobra.Command{
		Use:   "scoring-bands",
		Short: "View scoring bands",
		Long:  `View scoring bands used for calculating scores.`,
		Run:   runViewScoringBandsCmd,
	}

	categoryWeightsCmd := &cobra.Command{
		Use:   "category-weights",
		Short: "View category weights",
		Long:  `View weights assigned to each category for overall score calculation.`,
		Run:   runViewCategoryWeightsCmd,
	}

	categoryThresholdsCmd := &cobra.Command{
		Use:   "category-thresholds",
		Short: "View category-specific thresholds",
		Long:  `View category-specific thresholds for the traffic light model.`,
		Run:   runViewCategoryThresholdsCmd,
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate all lever configurations",
		Long:  `Validate both category weights and threshold configurations.`,
		Run:   runValidateLeversCmd,
	}

	validateWeightsCmd := &cobra.Command{
		Use:   "validate-weights",
		Short: "Validate category weights",
		Long:  `Validate that category weights add up to 100%.`,
		Run:   runValidateWeightsCmd,
	}

	validateThresholdsCmd := &cobra.Command{
		Use:   "validate-thresholds",
		Short: "Validate threshold configurations",
		Long:  `Validate that global and category-specific thresholds do not overlap and follow the correct order (Red < Yellow < Green).`,
		Run:   runValidateThresholdsCmd,
	}

	// Add subcommands to levers command
	leversCmd.AddCommand(allLeversCmd, globalThresholdsCmd, scoringBandsCmd, categoryWeightsCmd, categoryThresholdsCmd, validateCmd, validateWeightsCmd, validateThresholdsCmd)

	// Add metrics subcommand
	metricsCmd := &cobra.Command{
		Use:   "metrics",
		Short: "Manage metrics and metric files",
		Long:  `Commands for managing metrics and metric files.`,
	}

	// Add list metrics subcommand
	listMetricsCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available metrics",
		Long:  `List all available metrics with their current values.`,
		Run:   runListMetricsCmd,
	}

	// Add list-files subcommand
	listFilesCmd := &cobra.Command{
		Use:   "list-files",
		Short: "List all metric files",
		Long:  `List all metric files in the metrics directory.`,
		Run:   runListMetricFilesCmd,
	}

	// Add create-file subcommand
	createFileCmd := &cobra.Command{
		Use:   "create-file [name]",
		Short: "Create a new metric file",
		Long:  `Create a new empty metric file with the given name.`,
		Args:  cobra.ExactArgs(1),
		Run:   runCreateMetricFileCmd,
	}

	// Add subcommands to metrics command
	metricsCmd.AddCommand(listMetricsCmd, listFilesCmd, createFileCmd)

	// Add categories subcommand
	categoriesCmd := &cobra.Command{
		Use:   "categories",
		Short: "List all available categories",
		Long:  `List all available categories with their KPIs and KRIs.`,
		Run:   runListCategoriesCmd,
	}

	// Add init command
	initCmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialize configuration files",
		Long:  `Create default configuration files in the specified directory. If no directory is provided, files will be created in the default location (~/.pulse/).`,
		Args:  cobra.MaximumNArgs(1),
		Run:   runInitCmd,
	}

	// Add subcommands to list command
	listCmd.AddCommand(categoriesCmd)

	// Add version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version and build information of the Pulse CLI.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Pulse CLI version %s+%s\n", version, build)
		},
	}

	// Add commands to root command
	rootCmd.AddCommand(reportCmd, updateCmd, listCmd, metricsCmd, leversCmd, initCmd, versionCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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

func runListMetricFilesCmd(cmd *cobra.Command, args []string) {
	// Get metrics directory
	metricsDir := filepath.Join(dataDir, "metrics")

	// Check if metrics directory exists
	if _, err := os.Stat(metricsDir); os.IsNotExist(err) {
		fmt.Println("Metrics directory does not exist.")
		return
	}

	// Read all files in the metrics directory
	files, err := os.ReadDir(metricsDir)
	if err != nil {
		fmt.Printf("Error reading metrics directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No metric files found.")
		return
	}

	fmt.Println("Available metric files:")
	fmt.Println("----------------------")

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
			fmt.Println(file.Name())
		}
	}

	// Check for legacy file
	legacyPath := filepath.Join(dataDir, "metrics.yaml")
	if _, err := os.Stat(legacyPath); err == nil {
		fmt.Println("metrics.yaml (legacy format)")
	}
}

func runCreateMetricFileCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Create the metric file
	fileName := args[0]
	err := configLoader.CreateMetricFile(fileName)
	if err != nil {
		fmt.Printf("Error creating metric file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Metric file '%s' created successfully.\n", fileName)
}

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

func runInitCmd(cmd *cobra.Command, args []string) {
	var targetConfigDir, targetDataDir string

	if len(args) > 0 {
		// Use the specified directory
		targetDir := args[0]

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(targetDir, 0700); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", targetDir, err)
			os.Exit(1)
		}

		targetConfigDir = filepath.Join(targetDir, "config")
		targetDataDir = filepath.Join(targetDir, "data")
	} else {
		// Use the default directories
		targetConfigDir = configDir
		targetDataDir = dataDir
	}

	// Initialize the config loader with the target directories
	configLoader := pulse.NewConfigLoader(targetConfigDir, targetDataDir)

	// Create default configuration files
	err := configLoader.CreateDefaultConfigFiles()
	if err != nil {
		fmt.Printf("Error creating default configuration files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Default configuration files created in:\n")
	fmt.Printf("  Config directory: %s\n", targetConfigDir)
	fmt.Printf("  Data directory: %s\n", targetDataDir)
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

// runViewGlobalThresholdsCmd displays global thresholds
func runViewGlobalThresholdsCmd(cmd *cobra.Command, args []string) {
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
}

// runViewScoringBandsCmd displays scoring bands
func runViewScoringBandsCmd(cmd *cobra.Command, args []string) {
	// Note: We don't need to load any configs since we're just showing the concept

	fmt.Println("Scoring Bands:")
	fmt.Println("--------------")
	fmt.Println("Scoring bands are now defined per-metric with min/max ranges and scores.")
	fmt.Println("To view specific scoring bands, check the metrics configuration.")
	fmt.Println()

	// Display example of the new scoring band structure
	fmt.Println("Example of new scoring band structure:")
	fmt.Println("  - score: 100")
	fmt.Println("    min: 95")
	fmt.Println("  - score: 85")
	fmt.Println("    max: 94")
	fmt.Println("    min: 80")
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

// runViewCategoryThresholdsCmd displays category-specific thresholds
func runViewCategoryThresholdsCmd(cmd *cobra.Command, args []string) {
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Load levers configuration
	leversConfig, err := configLoader.LoadLeversConfig()
	if err != nil {
		fmt.Printf("Error loading levers config: %v\n", err)
		os.Exit(1)
	}

	// Display category-specific thresholds
	fmt.Println("Category-Specific Thresholds:")
	fmt.Println("----------------------------")
	if len(leversConfig.Weights.CategoryThresholds) == 0 {
		fmt.Println("No category-specific thresholds defined.")
	} else {
		for category, thresholds := range leversConfig.Weights.CategoryThresholds {
			fmt.Printf("%s:\n", category)
			fmt.Printf("  Green:  %d-%d\n", thresholds.Green.Min, thresholds.Green.Max)
			fmt.Printf("  Yellow: %d-%d\n", thresholds.Yellow.Min, thresholds.Yellow.Max)
			fmt.Printf("  Red:    %d-%d\n", thresholds.Red.Min, thresholds.Red.Max)
		}
	}
}

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
