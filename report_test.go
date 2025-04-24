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
						ScoringBands: map[string]int{
							"band_5": 5,
							"band_4": 10,
							"band_3": 15,
							"band_2": 20,
							"band_1": 21,
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
						ScoringBands: map[string]int{
							"band_5": 0,
							"band_4": 2,
							"band_3": 5,
							"band_2": 10,
							"band_1": 11,
						},
					},
				},
			},
		},
	}

	leversConfig := &LeversConfig{
		Global: Global{
			Thresholds: Thresholds{
				Green:  80,
				Yellow: 60,
				Red:    0,
			},
			ScoringBands: ScoringBands{
				Band5: 90,
				Band4: 80,
				Band3: 70,
				Band2: 60,
				Band1: 0,
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

	// Create a ReportGenerator
	generator := NewReportGenerator(calculator)

	// Test GenerateOverallReport with TextFormat
	textReport, err := generator.GenerateOverallReport(TextFormat)
	if err != nil {
		t.Fatalf("Failed to generate overall text report: %v", err)
	}

	// Check if the text report contains expected content
	expectedTextContent := []string{
		"SECURITY POSTURE REPORT",
		"Overall Score:",
		"Test Category:",
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
	if _, ok := jsonData["overall_score"]; !ok {
		t.Error("Expected JSON report to contain 'overall_score' field")
	}
	if _, ok := jsonData["status"]; !ok {
		t.Error("Expected JSON report to contain 'status' field")
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
		"Category Score:",
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
	if _, ok := categoryJsonData["category_score"]; !ok {
		t.Error("Expected category JSON report to contain 'category_score' field")
	}
	if _, ok := categoryJsonData["status"]; !ok {
		t.Error("Expected category JSON report to contain 'status' field")
	}
	if _, ok := categoryJsonData["metrics"]; !ok {
		t.Error("Expected category JSON report to contain 'metrics' field")
	}

	// Test GenerateCategoryReport with non-existent category
	_, err = generator.GenerateCategoryReport("non_existent", TextFormat)
	if err == nil {
		t.Error("Expected error for non-existent category, got nil")
	}

	// Test formatStatus
	if formatStatus(Green) != "GREEN" {
		t.Errorf("Expected formatStatus(Green) to be 'GREEN', got '%s'", formatStatus(Green))
	}
	if formatStatus(Yellow) != "YELLOW" {
		t.Errorf("Expected formatStatus(Yellow) to be 'YELLOW', got '%s'", formatStatus(Yellow))
	}
	if formatStatus(Red) != "RED" {
		t.Errorf("Expected formatStatus(Red) to be 'RED', got '%s'", formatStatus(Red))
	}
	if formatStatus("unknown") != "UNKNOWN" {
		t.Errorf("Expected formatStatus('unknown') to be 'UNKNOWN', got '%s'", formatStatus("unknown"))
	}
}
