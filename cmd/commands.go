package main

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
