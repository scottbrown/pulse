package pulse

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed defaults/config/* defaults/data/metrics/*
var defaultFiles embed.FS

// MigrateMetricsData migrates metrics from both legacy formats to the new format
func (c *ConfigLoader) MigrateMetricsData() error {
	allMetrics := &MetricsData{
		Metrics: []Metric{},
	}

	// 1. Check and migrate legacy file if it exists
	legacyPath := filepath.Join(c.DataDir, "metrics.yaml")
	if _, err := os.Stat(legacyPath); err == nil {
		// Legacy file exists, read it
		data, err := os.ReadFile(legacyPath)
		if err != nil {
			return fmt.Errorf("failed to read legacy metrics file: %w", err)
		}

		// Validate YAML before parsing
		if err := validateYAML(data); err != nil {
			return fmt.Errorf("invalid legacy metrics file: %w", err)
		}

		var legacyMetrics MetricsData
		if err := yaml.Unmarshal(data, &legacyMetrics); err != nil {
			return fmt.Errorf("failed to parse legacy metrics file: %w", err)
		}

		// Add source file information to each metric
		for i := range legacyMetrics.Metrics {
			parts := strings.Split(legacyMetrics.Metrics[i].Reference, ".")
			if len(parts) >= 1 {
				categoryID := parts[0]
				legacyMetrics.Metrics[i].SourceFile = categoryID + ".yaml"
			} else {
				legacyMetrics.Metrics[i].SourceFile = "default.yaml"
			}
		}

		// Add to all metrics
		allMetrics.Metrics = append(allMetrics.Metrics, legacyMetrics.Metrics...)

		// Rename legacy file to .bak
		backupPath := legacyPath + ".bak"
		if err := os.Rename(legacyPath, backupPath); err != nil {
			return fmt.Errorf("failed to rename legacy metrics file: %w", err)
		}
	}

	// 2. Check and migrate metrics directory if it exists
	metricsDir := filepath.Join(c.DataDir, "metrics")
	if _, err := os.Stat(metricsDir); err == nil {
		// Metrics directory exists, read all files
		files, err := os.ReadDir(metricsDir)
		if err != nil {
			return fmt.Errorf("failed to read metrics directory: %w", err)
		}

		// Process each YAML file in the directory
		for _, file := range files {
			if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") &&
				!strings.HasSuffix(file.Name(), ".yml")) {
				continue
			}

			path := filepath.Join(metricsDir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read metrics file %s: %w", file.Name(), err)
			}

			// Validate YAML before parsing
			if err := validateYAML(data); err != nil {
				return fmt.Errorf("invalid metrics file %s: %w", file.Name(), err)
			}

			var fileMetrics MetricsData
			if err := yaml.Unmarshal(data, &fileMetrics); err != nil {
				return fmt.Errorf("failed to parse metrics file %s: %w", file.Name(), err)
			}

			// Add source file information to each metric
			for i := range fileMetrics.Metrics {
				parts := strings.Split(fileMetrics.Metrics[i].Reference, ".")
				if len(parts) >= 1 {
					categoryID := parts[0]
					fileMetrics.Metrics[i].SourceFile = categoryID + ".yaml"
				} else {
					fileMetrics.Metrics[i].SourceFile = file.Name()
				}
			}

			// Add to all metrics
			allMetrics.Metrics = append(allMetrics.Metrics, fileMetrics.Metrics...)
		}

		// Rename metrics directory to .bak
		backupDir := metricsDir + ".bak"
		if err := os.Rename(metricsDir, backupDir); err != nil {
			return fmt.Errorf("failed to rename metrics directory: %w", err)
		}
	}

	// If we have metrics to migrate, save them to the new format
	if len(allMetrics.Metrics) > 0 {
		// Group metrics by category
		metricsByCategory := make(map[string][]Metric)

		for _, metric := range allMetrics.Metrics {
			sourceFile := metric.SourceFile
			if sourceFile == "" {
				parts := strings.Split(metric.Reference, ".")
				if len(parts) >= 1 {
					sourceFile = parts[0] + ".yaml"
				} else {
					sourceFile = "default.yaml"
				}
			}
			metricsByCategory[sourceFile] = append(metricsByCategory[sourceFile], metric)
		}

		// Save metrics to new files
		for fileName, metrics := range metricsByCategory {
			filePath := filepath.Join(c.DataDir, fileName)

			fileData := MetricsData{
				Metrics: metrics,
			}

			data, err := yaml.Marshal(fileData)
			if err != nil {
				return fmt.Errorf("failed to marshal metrics data for %s: %w", fileName, err)
			}

			if err := os.WriteFile(filePath, data, 0600); err != nil {
				return fmt.Errorf("failed to write metrics file %s: %w", fileName, err)
			}
		}
	}

	return nil
}

// Global mutex for file operations
var fileLock sync.Mutex

// ConfigLoader handles loading and parsing of configuration files
type ConfigLoader struct {
	ConfigDir string
	DataDir   string
}

// NewConfigLoader creates a new ConfigLoader with the specified directories
func NewConfigLoader(configDir, dataDir string) *ConfigLoader {
	return &ConfigLoader{
		ConfigDir: configDir,
		DataDir:   dataDir,
	}
}

// validateYAML performs basic validation on YAML data before parsing
func validateYAML(data []byte) error {
	// Check for YAML document markers that could indicate custom types
	if bytes.Contains(data, []byte("!!")) {
		return fmt.Errorf("potentially unsafe YAML: custom type tags detected")
	}

	// Check for YAML anchors and aliases that could be used for exploits
	// Only detect actual YAML anchors/aliases, not & or * in text
	// YAML anchors look like "&anchor_name" at the start of a line or after a colon
	// YAML aliases look like "*alias_name" at the start of a line or after a colon
	anchorPattern := []byte("\n&")
	aliasPattern := []byte("\n*")
	colonAnchorPattern := []byte(": &")
	colonAliasPattern := []byte(": *")

	if bytes.Contains(data, anchorPattern) ||
		bytes.Contains(data, aliasPattern) ||
		bytes.Contains(data, colonAnchorPattern) ||
		bytes.Contains(data, colonAliasPattern) {
		return fmt.Errorf("potentially unsafe YAML: anchors or aliases detected")
	}

	// Check for excessively large files
	if len(data) > 10*1024*1024 { // 10MB limit
		return fmt.Errorf("YAML file too large: %d bytes", len(data))
	}

	return nil
}

// LoadMetricsConfig loads the metrics configuration from the YAML file
func (c *ConfigLoader) LoadMetricsConfig() (*MetricsConfig, error) {
	path := filepath.Join(c.ConfigDir, "metrics.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		// If the file doesn't exist, return an empty config instead of an error
		if os.IsNotExist(err) {
			return &MetricsConfig{
				Categories: []Category{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read metrics config file: %w", err)
	}

	// Validate YAML before parsing
	if err := validateYAML(data); err != nil {
		return nil, fmt.Errorf("invalid metrics config file: %w", err)
	}

	var config MetricsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse metrics config file: %w", err)
	}

	return &config, nil
}

// LoadLeversConfig loads the executive levers configuration from the YAML file
func (c *ConfigLoader) LoadLeversConfig() (*LeversConfig, error) {
	path := filepath.Join(c.ConfigDir, "levers.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		// If the file doesn't exist, return an empty config instead of an error
		if os.IsNotExist(err) {
			return &LeversConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read levers config file: %w", err)
	}

	// Validate YAML before parsing
	if err := validateYAML(data); err != nil {
		return nil, fmt.Errorf("invalid levers config file: %w", err)
	}

	var config LeversConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse levers config file: %w", err)
	}

	return &config, nil
}

// LoadMetricsData loads the metrics data from YAML files in the data directory
func (c *ConfigLoader) LoadMetricsData() (*MetricsData, error) {
	// Check if data directory exists
	if _, err := os.Stat(c.DataDir); os.IsNotExist(err) {
		return &MetricsData{
			Metrics: []Metric{},
		}, nil
	}

	// Read all files in the data directory
	files, err := os.ReadDir(c.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	allMetrics := &MetricsData{
		Metrics: []Metric{},
	}

	// If no files found, return empty metrics
	if len(files) == 0 {
		return allMetrics, nil
	}

	// Process each YAML file in the directory
	var parseErrors []string

	for _, file := range files {
		// Skip directories and non-YAML files
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") &&
			!strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		path := filepath.Join(c.DataDir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("failed to read metrics file %s: %v", file.Name(), err))
			continue
		}

		// Validate YAML before parsing
		if err := validateYAML(data); err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("invalid metrics file %s: %v", file.Name(), err))
			continue
		}

		var fileMetrics MetricsData
		if err := yaml.Unmarshal(data, &fileMetrics); err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("failed to parse metrics file %s: %v", file.Name(), err))
			continue
		}

		// Add source file information to each metric for saving later
		for i := range fileMetrics.Metrics {
			fileMetrics.Metrics[i].SourceFile = file.Name()
		}

		allMetrics.Metrics = append(allMetrics.Metrics, fileMetrics.Metrics...)
	}

	// If we couldn't parse any files and have errors, return the first error
	if len(allMetrics.Metrics) == 0 && len(parseErrors) > 0 {
		return nil, fmt.Errorf("failed to load any metrics: %s", parseErrors[0])
	}

	// If we have some metrics, return them even if there were some errors
	return allMetrics, nil
}

// SaveMetricsData saves the metrics data to YAML files in the data directory
func (c *ConfigLoader) SaveMetricsData(metricsData *MetricsData) error {
	// Use global mutex to prevent concurrent access to file operations
	fileLock.Lock()
	defer fileLock.Unlock()

	// Ensure data directory exists
	if err := os.MkdirAll(c.DataDir, 0700); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Group metrics by source file
	metricsByFile := make(map[string][]Metric)

	for _, metric := range metricsData.Metrics {
		sourceFile := metric.SourceFile
		if sourceFile == "" {
			// If no source file specified, use default
			sourceFile = "default.yaml"
		}
		metricsByFile[sourceFile] = append(metricsByFile[sourceFile], metric)
	}

	// Save each group to its respective file
	for fileName, metrics := range metricsByFile {
		filePath := filepath.Join(c.DataDir, fileName)

		fileData := MetricsData{
			Metrics: metrics,
		}

		data, err := yaml.Marshal(fileData)
		if err != nil {
			return fmt.Errorf("failed to marshal metrics data for %s: %w", fileName, err)
		}

		// Use atomic file write pattern
		tempFile := filePath + ".tmp"
		if err := os.WriteFile(tempFile, data, 0600); err != nil {
			return fmt.Errorf("failed to write temporary metrics data file %s: %w", fileName, err)
		}

		if err := os.Rename(tempFile, filePath); err != nil {
			// Try to clean up the temp file
			os.Remove(tempFile)
			return fmt.Errorf("failed to rename temporary metrics data file %s: %w", fileName, err)
		}
	}

	return nil
}

// CreateMetricFile creates a new metric file with the given name
func (c *ConfigLoader) CreateMetricFile(fileName string) error {
	// Use global mutex to prevent concurrent access to file operations
	fileLock.Lock()
	defer fileLock.Unlock()

	if !strings.HasSuffix(fileName, ".yaml") && !strings.HasSuffix(fileName, ".yml") {
		fileName += ".yaml"
	}

	// Validate filename to prevent path traversal
	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		return fmt.Errorf("invalid filename: %s (contains path separators)", fileName)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(c.DataDir, 0700); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	filePath := filepath.Join(c.DataDir, fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("metric file %s already exists", fileName)
	}

	// Create empty metrics file
	emptyMetrics := MetricsData{
		Metrics: []Metric{},
	}

	data, err := yaml.Marshal(emptyMetrics)
	if err != nil {
		return fmt.Errorf("failed to marshal empty metrics data: %w", err)
	}

	// Use atomic file write pattern
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary metric file %s: %w", fileName, err)
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		// Try to clean up the temp file
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temporary metric file %s: %w", fileName, err)
	}

	return nil
}

// CreateDefaultConfigFiles creates default configuration files if they don't exist
func (c *ConfigLoader) CreateDefaultConfigFiles() error {
	// Ensure directories exist
	if err := os.MkdirAll(c.ConfigDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.MkdirAll(c.DataDir, 0700); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Helper function to copy embedded file to destination
	copyEmbeddedFile := func(embeddedPath, destPath string) error {
		// Check if destination file already exists
		if _, err := os.Stat(destPath); err == nil {
			// File exists, skip
			return nil
		}

		// Read embedded file
		data, err := defaultFiles.ReadFile(embeddedPath)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", embeddedPath, err)
		}

		// Write to destination
		if err := os.WriteFile(destPath, data, 0600); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}

		return nil
	}

	// Copy config files
	if err := copyEmbeddedFile("defaults/config/metrics.yaml", filepath.Join(c.ConfigDir, "metrics.yaml")); err != nil {
		return err
	}
	if err := copyEmbeddedFile("defaults/config/levers.yaml", filepath.Join(c.ConfigDir, "levers.yaml")); err != nil {
		return err
	}

	// Copy metrics data files directly to the data directory
	if err := copyEmbeddedFile("defaults/data/metrics/app_sec.yaml", filepath.Join(c.DataDir, "app_sec.yaml")); err != nil {
		return err
	}
	if err := copyEmbeddedFile("defaults/data/metrics/infra_sec.yaml", filepath.Join(c.DataDir, "infra_sec.yaml")); err != nil {
		return err
	}
	if err := copyEmbeddedFile("defaults/data/metrics/compliance.yaml", filepath.Join(c.DataDir, "compliance.yaml")); err != nil {
		return err
	}

	return nil
}
