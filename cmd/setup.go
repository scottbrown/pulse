package main

import (
	"fmt"
	"os"
	"path/filepath"
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

func setupDefaultDirs() (string, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("Error getting home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, defaultConfigDirName, "config")
	dataDir := filepath.Join(homeDir, defaultConfigDirName, "data")

	return configDir, dataDir, nil
}

func setupFlags(configDir, dataDir string) {
	// Add persistent flags for config and data directories
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", configDir, "Directory containing configuration files")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", dataDir, "Directory containing data files")

	reportCmd.Flags().StringVarP(&category, "category", "c", "", "Generate report for a specific category")
	reportCmd.Flags().StringVarP(&format, "format", "f", "text", "Report format (text or json)")
	reportCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	reportCmd.Flags().StringVar(&scoringMethod, "scoring-method", "median", "Scoring method to use (median or average)")

	updateCmd.Flags().StringVarP(&metricRef, "metric", "m", "", "Metric reference (e.g., app_sec.KPI.vuln_remediation_time)")
	updateCmd.Flags().StringVarP(&metricVal, "value", "v", "", "Metric value")
	updateCmd.MarkFlagRequired("metric")
	updateCmd.MarkFlagRequired("value")
}

func setupCommands() {
	// Add subcommands to levers command
	leversCmd.AddCommand(allLeversCmd, globalThresholdsCmd, scoringBandsCmd, categoryWeightsCmd, categoryThresholdsCmd, validateCmd, validateWeightsCmd, validateThresholdsCmd)

	// Add subcommands to metrics command
	metricsCmd.AddCommand(listMetricsCmd, listFilesCmd, createFileCmd)

	// Add subcommands to list command
	listCmd.AddCommand(categoriesCmd)

	// Add commands to root command
	rootCmd.AddCommand(reportCmd, updateCmd, listCmd, metricsCmd, leversCmd, initCmd, versionCmd)
}
