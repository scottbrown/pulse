package pulse

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoader(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "pulse-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configDir := filepath.Join(tempDir, "config")
	dataDir := filepath.Join(tempDir, "data")

	// Create a ConfigLoader instance
	loader := NewConfigLoader(configDir, dataDir)

	// Create test directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("Failed to create data directory: %v", err)
	}

	metricsDir := filepath.Join(dataDir, "metrics")
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		t.Fatalf("Failed to create metrics directory: %v", err)
	}

	// Create test config files
	metricsConfigData := `categories:
  - id: "test_cat"
    name: "Test Category"
    description: "Test category description"
    kpis:
      - id: "test_kpi"
        name: "Test KPI"
        description: "Test KPI description"
        unit: "count"
        target: 10
    kris:
      - id: "test_kri"
        name: "Test KRI"
        description: "Test KRI description"
        unit: "count"
        threshold: 5`

	if err := os.WriteFile(filepath.Join(configDir, "metrics.yaml"), []byte(metricsConfigData), 0644); err != nil {
		t.Fatalf("Failed to write metrics config file: %v", err)
	}

	leversConfigData := `global:
  thresholds:
    green:
      min: 80
      max: 100
    yellow:
      min: 60
      max: 79
    red:
      min: 0
      max: 59`

	if err := os.WriteFile(filepath.Join(configDir, "levers.yaml"), []byte(leversConfigData), 0644); err != nil {
		t.Fatalf("Failed to write levers config file: %v", err)
	}

	// Create test metrics file
	testMetricsData := `metrics:
  - reference: "test_cat.KPI.test_kpi"
    value: 10
    timestamp: "2025-04-01T00:00:00Z"
  - reference: "test_cat.KRI.test_kri"
    value: 5
    timestamp: "2025-04-01T00:00:00Z"`

	if err := os.WriteFile(filepath.Join(metricsDir, "test_cat.yaml"), []byte(testMetricsData), 0644); err != nil {
		t.Fatalf("Failed to write test metrics file: %v", err)
	}

	// Check if the files were created
	configFiles := []string{
		filepath.Join(configDir, "metrics.yaml"),
		filepath.Join(configDir, "levers.yaml"),
	}

	for _, file := range configFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}
	}

	// Check if metrics directory exists
	if _, err := os.Stat(metricsDir); os.IsNotExist(err) {
		t.Errorf("Expected metrics directory %s to exist", metricsDir)
	} else {
		// Check if metrics files exist
		metricFiles, err := os.ReadDir(metricsDir)
		if err != nil {
			t.Errorf("Failed to read metrics directory: %v", err)
		} else if len(metricFiles) == 0 {
			t.Errorf("Expected metrics files in %s, but directory is empty", metricsDir)
		}
	}

	// Test loading the created files
	metricsConfig, err := loader.LoadMetricsConfig()
	if err != nil {
		t.Fatalf("Failed to load metrics config: %v", err)
	}

	if len(metricsConfig.Categories) == 0 {
		t.Error("Expected categories in metrics config, got none")
	}

	leversConfig, err := loader.LoadLeversConfig()
	if err != nil {
		t.Fatalf("Failed to load levers config: %v", err)
	}

	if leversConfig.Global.Thresholds.Green.Min <= 0 {
		t.Error("Expected positive green threshold min value in levers config")
	}

	metricsData, err := loader.LoadMetricsData()
	if err != nil {
		t.Fatalf("Failed to load metrics data: %v", err)
	}

	if len(metricsData.Metrics) == 0 {
		t.Error("Expected metrics in metrics data, got none")
	}

	// Test SaveMetricsData
	newMetric := Metric{
		Reference:  "test.KPI.test_metric",
		Value:      42.0,
		SourceFile: "test.yaml",
	}
	metricsData.Metrics = append(metricsData.Metrics, newMetric)

	err = loader.SaveMetricsData(metricsData)
	if err != nil {
		t.Fatalf("Failed to save metrics data: %v", err)
	}

	// Load the data again to verify the save
	updatedData, err := loader.LoadMetricsData()
	if err != nil {
		t.Fatalf("Failed to load updated metrics data: %v", err)
	}

	found := false
	for _, metric := range updatedData.Metrics {
		if metric.Reference == "test.KPI.test_metric" && metric.Value == 42.0 {
			found = true
			break
		}
	}

	if !found {
		t.Error("Failed to find the newly added metric in the saved data")
	}
}
