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

	// Test CreateDefaultConfigFiles
	err = loader.CreateDefaultConfigFiles()
	if err != nil {
		t.Fatalf("Failed to create default config files: %v", err)
	}

	// Check if the files were created
	files := []string{
		filepath.Join(configDir, "metrics.yaml"),
		filepath.Join(configDir, "levers.yaml"),
		filepath.Join(dataDir, "metrics.yaml"),
	}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
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

	if leversConfig.Global.Thresholds.Green <= 0 {
		t.Error("Expected positive green threshold in levers config")
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
		Reference: "test.KPI.test_metric",
		Value:     42.0,
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
