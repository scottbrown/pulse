package main

import (
	"fmt"
	"os"
	"path/filepath"
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
