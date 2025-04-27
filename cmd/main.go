package main

import (
	"fmt"
	"os"
)

func main() {
	defaultConfigDir, defaultDataDir, err := setupDefaultDirs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	setupFlags(defaultConfigDir, defaultDataDir)
	setupCommands()

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
