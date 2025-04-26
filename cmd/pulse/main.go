package main

import (
	"fmt"
	"os"
	"path/filepath"

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
