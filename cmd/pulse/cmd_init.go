package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

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
