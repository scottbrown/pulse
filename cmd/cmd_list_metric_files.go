package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Add list-files subcommand
var listFilesCmd = &cobra.Command{
	Use:   "list-files",
	Short: "List all metric files",
	Long:  `List all metric files in the data directory.`,
	Run:   runListMetricFilesCmd,
}

func runListMetricFilesCmd(cmd *cobra.Command, args []string) {
	// Check if data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Println("Data directory does not exist.")
		return
	}

	// Read all files in the data directory
	files, err := os.ReadDir(dataDir)
	if err != nil {
		fmt.Printf("Error reading data directory: %v\n", err)
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
}
