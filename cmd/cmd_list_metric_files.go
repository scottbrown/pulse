package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Add list-files subcommand
var listFilesCmd = &cobra.Command{
	Use:   "list-files",
	Short: "List all metric files",
	Long:  `List all metric files in the metrics directory.`,
	Run:   runListMetricFilesCmd,
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
