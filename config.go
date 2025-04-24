package pulse

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

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
	if bytes.Contains(data, []byte("&")) || bytes.Contains(data, []byte("*")) {
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

// LoadMetricsData loads the metrics data from YAML files in the metrics directory
func (c *ConfigLoader) LoadMetricsData() (*MetricsData, error) {
	metricsDir := filepath.Join(c.DataDir, "metrics")

	// Check if metrics directory exists
	if _, err := os.Stat(metricsDir); os.IsNotExist(err) {
		// If not, try the legacy single file approach
		return c.loadLegacyMetricsData()
	}

	// Read all files in the metrics directory
	files, err := os.ReadDir(metricsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metrics directory: %w", err)
	}

	allMetrics := &MetricsData{
		Metrics: []Metric{},
	}

	// If no files found, try legacy approach
	if len(files) == 0 {
		legacyData, err := c.loadLegacyMetricsData()
		if err == nil {
			return legacyData, nil
		}
		// If legacy approach fails, return empty metrics
		return allMetrics, nil
	}

	// Process each YAML file in the directory
	var parseErrors []string

	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") &&
			!strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		path := filepath.Join(metricsDir, file.Name())
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

// loadLegacyMetricsData loads metrics from the legacy single file format
func (c *ConfigLoader) loadLegacyMetricsData() (*MetricsData, error) {
	path := filepath.Join(c.DataDir, "metrics.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metrics data file: %w", err)
	}

	// Validate YAML before parsing
	if err := validateYAML(data); err != nil {
		return nil, fmt.Errorf("invalid metrics data file: %w", err)
	}

	var metricsData MetricsData
	if err := yaml.Unmarshal(data, &metricsData); err != nil {
		return nil, fmt.Errorf("failed to parse metrics data file: %w", err)
	}

	// Mark all metrics as coming from the legacy file
	for i := range metricsData.Metrics {
		metricsData.Metrics[i].SourceFile = "metrics.yaml"
	}

	return &metricsData, nil
}

// SaveMetricsData saves the metrics data to YAML files in the metrics directory
func (c *ConfigLoader) SaveMetricsData(metricsData *MetricsData) error {
	// Use global mutex to prevent concurrent access to file operations
	fileLock.Lock()
	defer fileLock.Unlock()

	metricsDir := filepath.Join(c.DataDir, "metrics")

	// Ensure metrics directory exists
	if err := os.MkdirAll(metricsDir, 0700); err != nil {
		return fmt.Errorf("failed to create metrics directory: %w", err)
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
		filePath := filepath.Join(metricsDir, fileName)

		// For legacy file, use the original path
		if fileName == "metrics.yaml" {
			filePath = filepath.Join(c.DataDir, fileName)
		}

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

	metricsDir := filepath.Join(c.DataDir, "metrics")

	// Ensure metrics directory exists
	if err := os.MkdirAll(metricsDir, 0700); err != nil {
		return fmt.Errorf("failed to create metrics directory: %w", err)
	}

	filePath := filepath.Join(metricsDir, fileName)

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

	// Create default metrics config if it doesn't exist
	metricsConfigPath := filepath.Join(c.ConfigDir, "metrics.yaml")
	if _, err := os.Stat(metricsConfigPath); os.IsNotExist(err) {
		defaultMetricsConfig := `categories:
		- id: "app_sec"
		  name: "Application Security"
		  description: "Metrics related to application security posture"
		  kpis:
		    - id: "vuln_remediation_time"
		      name: "Vulnerability Remediation Time"
		      description: "Average time to remediate vulnerabilities"
		      unit: "days"
		      target: 30
		      scoring_bands:
		        - score: 95
		          max: 15
		        - score: 85
		          min: 15
		          max: 30
		        - score: 75
		          min: 30
		          max: 45
		        - score: 65
		          min: 45
		          max: 60
		        - score: 30
		          min: 60
		  kris:
		    - id: "critical_vulns"
		      name: "Critical Vulnerabilities"
		      description: "Number of critical vulnerabilities"
		      unit: "count"
		      threshold: 5
		      scoring_bands:
		        - score: 95
		          max: 0
		        - score: 85
		          min: 0
		          max: 2
		        - score: 75
		          min: 2
		          max: 5
		        - score: 65
		          min: 5
		          max: 10
		        - score: 30
		          min: 10`

		if err := os.WriteFile(metricsConfigPath, []byte(defaultMetricsConfig), 0600); err != nil {
			return fmt.Errorf("failed to create default metrics config: %w", err)
		}
	}

	// Create default levers config if it doesn't exist
	leversConfigPath := filepath.Join(c.ConfigDir, "levers.yaml")
	if _, err := os.Stat(leversConfigPath); os.IsNotExist(err) {
		defaultLeversConfig := `global:
		thresholds:
		  green:
		    min: 80
		    max: 100
		  yellow:
		    min: 60
		    max: 79
		  red:
		    min: 0
		    max: 59
		
weights:
		categories:
		  "app_sec": 0.4
		  "infra_sec": 0.3
		  "compliance": 0.3
		
		category_thresholds:
		  "compliance":
		    green:
		      min: 85
		      max: 100
		    yellow:
		      min: 70
		      max: 84
		    red:
		      min: 0
		      max: 69`

		if err := os.WriteFile(leversConfigPath, []byte(defaultLeversConfig), 0600); err != nil {
			return fmt.Errorf("failed to create default levers config: %w", err)
		}
	}

	// Create metrics directory and default metrics files
	metricsDir := filepath.Join(c.DataDir, "metrics")
	if err := os.MkdirAll(metricsDir, 0700); err != nil {
		return fmt.Errorf("failed to create metrics directory: %w", err)
	}

	// Create app_sec metrics file
	appSecPath := filepath.Join(metricsDir, "app_sec.yaml")
	if _, err := os.Stat(appSecPath); os.IsNotExist(err) {
		appSecData := `metrics:
		- reference: "app_sec.KPI.vuln_remediation_time"
		  value: 45
		  timestamp: "2025-04-01T00:00:00Z"
		- reference: "app_sec.KRI.critical_vulns"
		  value: 3
		  timestamp: "2025-04-01T00:00:00Z"`

		if err := os.WriteFile(appSecPath, []byte(appSecData), 0600); err != nil {
			return fmt.Errorf("failed to create app_sec metrics file: %w", err)
		}
	}

	// Create infra_sec metrics file
	infraSecPath := filepath.Join(metricsDir, "infra_sec.yaml")
	if _, err := os.Stat(infraSecPath); os.IsNotExist(err) {
		infraSecData := `metrics:
		- reference: "infra_sec.KPI.patch_coverage"
		  value: 94
		  timestamp: "2025-04-01T00:00:00Z"
		- reference: "infra_sec.KRI.exposed_services"
		  value: 4
		  timestamp: "2025-04-01T00:00:00Z"`

		if err := os.WriteFile(infraSecPath, []byte(infraSecData), 0600); err != nil {
			return fmt.Errorf("failed to create infra_sec metrics file: %w", err)
		}
	}

	// Create compliance metrics file
	compliancePath := filepath.Join(metricsDir, "compliance.yaml")
	if _, err := os.Stat(compliancePath); os.IsNotExist(err) {
		complianceData := `metrics:
		- reference: "compliance.KPI.policy_compliance"
		  value: 92
		  timestamp: "2025-04-01T00:00:00Z"
		- reference: "compliance.KRI.open_audit_findings"
		  value: 7
		  timestamp: "2025-04-01T00:00:00Z"`

		if err := os.WriteFile(compliancePath, []byte(complianceData), 0600); err != nil {
			return fmt.Errorf("failed to create compliance metrics file: %w", err)
		}
	}

	return nil
}
