package pulse

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestReportGenerator(t *testing.T) {
	// Create test data
	metricsConfig := &MetricsConfig{
		Categories: []Category{
			{
				ID:          "test_cat",
				Name:        "Test Category",
				Description: "Test category description",
				KPIs: []KPI{
					{
						ID:          "test_kpi",
						Name:        "Test KPI",
						Description: "Test KPI description",
						Unit:        "count",
						Target:      10,
						ScoringBands: []ScoringBand{
							{Max: FloatPtr(5), Score: 95},
							{Min: FloatPtr(5), Max: FloatPtr(10), Score: 85},
							{Min: FloatPtr(10), Max: FloatPtr(15), Score: 75},
							{Min: FloatPtr(15), Max: FloatPtr(20), Score: 65},
							{Min: FloatPtr(20), Score: 30},
						},
					},
				},
				KRIs: []KRI{
					{
						ID:          "test_kri",
						Name:        "Test KRI",
						Description: "Test KRI description",
						Unit:        "count",
						Threshold:   5,
						ScoringBands: []ScoringBand{
							{Max: FloatPtr(0), Score: 95},
							{Min: FloatPtr(0), Max: FloatPtr(2), Score: 85},
							{Min: FloatPtr(2), Max: FloatPtr(5), Score: 75},
							{Min: FloatPtr(5), Max: FloatPtr(10), Score: 65},
							{Min: FloatPtr(10), Score: 30},
						},
					},
				},
			},
		},
	}

	leversConfig := &LeversConfig{
		Global: Global{
			Thresholds: Thresholds{
				Green: ThresholdRange{
					Min: 80,
					Max: 100,
				},
				Yellow: ThresholdRange{
					Min: 60,
					Max: 79,
				},
				Red: ThresholdRange{
					Min: 0,
					Max: 59,
				},
			},
			KPIThresholds: Thresholds{
				Green: ThresholdRange{
					Min: 85,
					Max: 100,
				},
				Yellow: ThresholdRange{
					Min: 65,
					Max: 84,
				},
				Red: ThresholdRange{
					Min: 0,
					Max: 64,
				},
			},
			KRIThresholds: Thresholds{
				Green: ThresholdRange{
					Min: 75,
					Max: 100,
				},
				Yellow: ThresholdRange{
					Min: 55,
					Max: 74,
				},
				Red: ThresholdRange{
					Min: 0,
					Max: 54,
				},
			},
		},
		Weights: Weights{
			Categories: CategoryWeights{
				"test_cat": 1.0,
			},
		},
	}

	metricsData := &MetricsData{
		Metrics: []Metric{
			{
				Reference: "test_cat.KPI.test_kpi",
				Value:     3,
				Timestamp: time.Now(),
			},
			{
				Reference: "test_cat.KRI.test_kri",
				Value:     4,
				Timestamp: time.Now(),
			},
		},
	}

	// Create a MetricsProcessor
	processor := NewMetricsProcessor(metricsConfig, leversConfig, metricsData)

	// Create a ScoreCalculator with median scoring (default)
	calculator := NewScoreCalculator(processor, MedianScoring)

	// Create a ReportGenerator with text labels for testing
	generator := NewReportGenerator(calculator, TextLabels)

	// Test GenerateOverallReport with TextFormat
	textReport, err := generator.GenerateOverallReport(TextFormat)
	if err != nil {
		t.Fatalf("Failed to generate overall text report: %v", err)
	}

	// Check if the text report contains expected content
	expectedTextContent := []string{
		"SECURITY POSTURE REPORT",
		"KPI Score:",
		"KRI Score:",
		"Test Category:",
		"KPI: ",
		"KRI: ",
		"KPI test_kpi:",
		"KRI test_kri:",
	}

	for _, expected := range expectedTextContent {
		if !strings.Contains(textReport, expected) {
			t.Errorf("Expected text report to contain '%s', but it doesn't", expected)
		}
	}

	// Test GenerateOverallReport with JSONFormat
	jsonReport, err := generator.GenerateOverallReport(JSONFormat)
	if err != nil {
		t.Fatalf("Failed to generate overall JSON report: %v", err)
	}

	// Parse the JSON report
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonReport), &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON report: %v", err)
	}

	// Check if the JSON report contains expected fields
	if _, ok := jsonData["report_date"]; !ok {
		t.Error("Expected JSON report to contain 'report_date' field")
	}
	if _, ok := jsonData["kpi_score"]; !ok {
		t.Error("Expected JSON report to contain 'kpi_score' field")
	}
	if _, ok := jsonData["kri_score"]; !ok {
		t.Error("Expected JSON report to contain 'kri_score' field")
	}
	if _, ok := jsonData["kpi_status"]; !ok {
		t.Error("Expected JSON report to contain 'kpi_status' field")
	}
	if _, ok := jsonData["kri_status"]; !ok {
		t.Error("Expected JSON report to contain 'kri_status' field")
	}
	if _, ok := jsonData["categories"]; !ok {
		t.Error("Expected JSON report to contain 'categories' field")
	}

	// Test GenerateCategoryReport with TextFormat
	categoryTextReport, err := generator.GenerateCategoryReport("test_cat", TextFormat)
	if err != nil {
		t.Fatalf("Failed to generate category text report: %v", err)
	}

	// Check if the category text report contains expected content
	expectedCategoryTextContent := []string{
		"TEST CATEGORY REPORT",
		"KPI Score:",
		"KRI Score:",
		"KPI test_kpi:",
		"KRI test_kri:",
	}

	for _, expected := range expectedCategoryTextContent {
		if !strings.Contains(categoryTextReport, expected) {
			t.Errorf("Expected category text report to contain '%s', but it doesn't", expected)
		}
	}

	// Test GenerateCategoryReport with JSONFormat
	categoryJsonReport, err := generator.GenerateCategoryReport("test_cat", JSONFormat)
	if err != nil {
		t.Fatalf("Failed to generate category JSON report: %v", err)
	}

	// Parse the category JSON report
	var categoryJsonData map[string]interface{}
	if err := json.Unmarshal([]byte(categoryJsonReport), &categoryJsonData); err != nil {
		t.Fatalf("Failed to parse category JSON report: %v", err)
	}

	// Check if the category JSON report contains expected fields
	if _, ok := categoryJsonData["report_date"]; !ok {
		t.Error("Expected category JSON report to contain 'report_date' field")
	}
	if _, ok := categoryJsonData["category_id"]; !ok {
		t.Error("Expected category JSON report to contain 'category_id' field")
	}
	if _, ok := categoryJsonData["category_name"]; !ok {
		t.Error("Expected category JSON report to contain 'category_name' field")
	}
	if _, ok := categoryJsonData["kpi_score"]; !ok {
		t.Error("Expected category JSON report to contain 'kpi_score' field")
	}
	if _, ok := categoryJsonData["kri_score"]; !ok {
		t.Error("Expected category JSON report to contain 'kri_score' field")
	}
	if _, ok := categoryJsonData["kpi_status"]; !ok {
		t.Error("Expected category JSON report to contain 'kpi_status' field")
	}
	if _, ok := categoryJsonData["kri_status"]; !ok {
		t.Error("Expected category JSON report to contain 'kri_status' field")
	}
	if _, ok := categoryJsonData["metrics"]; !ok {
		t.Error("Expected category JSON report to contain 'metrics' field")
	}

	// Test GenerateCategoryReport with non-existent category
	_, err = generator.GenerateCategoryReport("non_existent", TextFormat)
	if err == nil {
		t.Error("Expected error for non-existent category, got nil")
	}

	// Test formatStatus with text labels
	if generator.formatStatus(Green) != "GREEN" {
		t.Errorf("Expected generator.formatStatus(Green) to be 'GREEN', got '%s'", generator.formatStatus(Green))
	}
	if generator.formatStatus(Yellow) != "YELLOW" {
		t.Errorf("Expected generator.formatStatus(Yellow) to be 'YELLOW', got '%s'", generator.formatStatus(Yellow))
	}
	if generator.formatStatus(Red) != "RED" {
		t.Errorf("Expected generator.formatStatus(Red) to be 'RED', got '%s'", generator.formatStatus(Red))
	}
	if generator.formatStatus("unknown") != "UNKNOWN" {
		t.Errorf("Expected generator.formatStatus('unknown') to be 'UNKNOWN', got '%s'", generator.formatStatus("unknown"))
	}

	// Create a ReportGenerator with emoji labels for testing
	emojiGenerator := NewReportGenerator(calculator, EmojiLabels)

	// Test formatStatus with emoji labels
	if emojiGenerator.formatStatus(Green) != "🟢" {
		t.Errorf("Expected emojiGenerator.formatStatus(Green) to be '🟢', got '%s'", emojiGenerator.formatStatus(Green))
	}
	if emojiGenerator.formatStatus(Yellow) != "🟡" {
		t.Errorf("Expected emojiGenerator.formatStatus(Yellow) to be '🟡', got '%s'", emojiGenerator.formatStatus(Yellow))
	}
	if emojiGenerator.formatStatus(Red) != "🔴" {
		t.Errorf("Expected emojiGenerator.formatStatus(Red) to be '🔴', got '%s'", emojiGenerator.formatStatus(Red))
	}
	if emojiGenerator.formatStatus("unknown") != "❓" {
		t.Errorf("Expected emojiGenerator.formatStatus('unknown') to be '❓', got '%s'", emojiGenerator.formatStatus("unknown"))
	}
}
