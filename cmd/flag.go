package main

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

func setupFlags(defaultConfigDir, defaultDataDir string) {
	// Add persistent flags for config and data directories
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", defaultConfigDir, "Directory containing configuration files")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", defaultDataDir, "Directory containing data files")

	reportCmd.Flags().StringVarP(&category, "category", "c", "", "Generate report for a specific category")
	reportCmd.Flags().StringVarP(&format, "format", "f", "text", "Report format (text or json)")
	reportCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	reportCmd.Flags().StringVar(&scoringMethod, "scoring-method", "median", "Scoring method to use (median or average)")

	updateCmd.Flags().StringVarP(&metricRef, "metric", "m", "", "Metric reference (e.g., app_sec.KPI.vuln_remediation_time)")
	updateCmd.Flags().StringVarP(&metricVal, "value", "v", "", "Metric value")
	updateCmd.MarkFlagRequired("metric")
	updateCmd.MarkFlagRequired("value")
}
