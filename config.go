package pulse

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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

// LoadMetricsConfig loads the metrics configuration from the YAML file
func (c *ConfigLoader) LoadMetricsConfig() (*MetricsConfig, error) {
	path := filepath.Join(c.ConfigDir, "metrics.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metrics config file: %w", err)
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

	var config LeversConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse levers config file: %w", err)
	}

	return &config, nil
}

// LoadMetricsData loads the metrics data from the YAML file
func (c *ConfigLoader) LoadMetricsData() (*MetricsData, error) {
	path := filepath.Join(c.DataDir, "metrics.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metrics data file: %w", err)
	}

	var metricsData MetricsData
	if err := yaml.Unmarshal(data, &metricsData); err != nil {
		return nil, fmt.Errorf("failed to parse metrics data file: %w", err)
	}

	return &metricsData, nil
}

// SaveMetricsData saves the metrics data to the YAML file
func (c *ConfigLoader) SaveMetricsData(metricsData *MetricsData) error {
	path := filepath.Join(c.DataDir, "metrics.yaml")

	data, err := yaml.Marshal(metricsData)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics data: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics data file: %w", err)
	}

	return nil
}

// CreateDefaultConfigFiles creates default configuration files if they don't exist
func (c *ConfigLoader) CreateDefaultConfigFiles() error {
	// Ensure directories exist
	if err := os.MkdirAll(c.ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.MkdirAll(c.DataDir, 0755); err != nil {
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
          band_5: 15
          band_4: 30
          band_3: 45
          band_2: 60
          band_1: 61
    kris:
      - id: "critical_vulns"
        name: "Critical Vulnerabilities"
        description: "Number of critical vulnerabilities"
        unit: "count"
        threshold: 5
        scoring_bands:
          band_5: 0
          band_4: 2
          band_3: 5
          band_2: 10
          band_1: 11`

		if err := os.WriteFile(metricsConfigPath, []byte(defaultMetricsConfig), 0644); err != nil {
			return fmt.Errorf("failed to create default metrics config: %w", err)
		}
	}

	// Create default levers config if it doesn't exist
	leversConfigPath := filepath.Join(c.ConfigDir, "levers.yaml")
	if _, err := os.Stat(leversConfigPath); os.IsNotExist(err) {
		defaultLeversConfig := `global:
  thresholds:
    green: 80
    yellow: 60
    red: 0
  
  scoring_bands:
    band_5: 90
    band_4: 80
    band_3: 70
    band_2: 60
    band_1: 0
  
weights:
  categories:
    "app_sec": 0.4
    "infra_sec": 0.3
    "compliance": 0.3
  
  category_thresholds:
    "compliance":
      green: 85
      yellow: 70
      red: 0`

		if err := os.WriteFile(leversConfigPath, []byte(defaultLeversConfig), 0644); err != nil {
			return fmt.Errorf("failed to create default levers config: %w", err)
		}
	}

	// Create default metrics data if it doesn't exist
	metricsDataPath := filepath.Join(c.DataDir, "metrics.yaml")
	if _, err := os.Stat(metricsDataPath); os.IsNotExist(err) {
		defaultMetricsData := `metrics:
  - reference: "app_sec.KPI.vuln_remediation_time"
    value: 45
    timestamp: "2025-04-01T00:00:00Z"
  - reference: "app_sec.KRI.critical_vulns"
    value: 3
    timestamp: "2025-04-01T00:00:00Z"`

		if err := os.WriteFile(metricsDataPath, []byte(defaultMetricsData), 0644); err != nil {
			return fmt.Errorf("failed to create default metrics data: %w", err)
		}
	}

	return nil
}
