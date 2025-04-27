package main

import (
	"fmt"

	"github.com/scottbrown/pulse"
	"github.com/spf13/cobra"
)

// Define root command
var rootCmd = &cobra.Command{
	Use:     pulse.ApplicationName,
	Short:   "Pulse - Risk and Performance measurement framework CLI",
	Long:    `A CLI application for reporting on Key Performance Indicators (KPIs) and Key Risk Indicators (KRIs) for organizational programs.`,
	Version: fmt.Sprintf("%s (%s)", pulse.Version, pulse.Build),
}
