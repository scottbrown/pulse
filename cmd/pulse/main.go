package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// Version information set by build flags
var (
	version = "main"
	build   = "unknown"
)

var (
	configDir  string
	dataDir    string
	category   string
	format     string
	outputFile string
	metricRef  string
	metricVal  string
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

	// Add metrics subcommand
	metricsCmd := &cobra.Command{
		Use:   "metrics",
		Short: "List all available metrics",
		Long:  `List all available metrics with their current values.`,
		Run:   runListMetricsCmd,
	}

	// Add categories subcommand
	categoriesCmd := &cobra.Command{
		Use:   "categories",
		Short: "List all available categories",
		Long:  `List all available categories with their KPIs and KRIs.`,
		Run:   runListCategoriesCmd,
	}

	// Add init command
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration files",
		Long:  `Create default configuration files if they don't exist.`,
		Run:   runInitCmd,
	}

	// Add subcommands to list command
	listCmd.AddCommand(metricsCmd, categoriesCmd)

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
	rootCmd.AddCommand(reportCmd, updateCmd, listCmd, initCmd, versionCmd)

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

	// Initialize the score calculator
	scoreCalculator := pulse.NewScoreCalculator(metricsProcessor)

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
		err := os.WriteFile(outputFile, []byte(reportContent), 0644)
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

	// Parse the metric value
	value, err := strconv.ParseFloat(metricVal, 64)
	if err != nil {
		fmt.Printf("Error parsing metric value: %v\n", err)
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
	// Initialize the config loader
	configLoader := pulse.NewConfigLoader(configDir, dataDir)

	// Create default configuration files
	err := configLoader.CreateDefaultConfigFiles()
	if err != nil {
		fmt.Printf("Error creating default configuration files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Default configuration files created in:\n")
	fmt.Printf("  Config directory: %s\n", configDir)
	fmt.Printf("  Data directory: %s\n", dataDir)
}
