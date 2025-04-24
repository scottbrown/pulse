package main

import (
	"fmt"
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
	rootCmd.AddCommand(reportCmd, updateCmd, listCmd, metricsCmd, initCmd, versionCmd)

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
		if err := os.MkdirAll(targetDir, 0755); err != nil {
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
