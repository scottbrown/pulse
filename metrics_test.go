package pulse

import (
	"testing"
	"time"
)

func TestMetricsProcessor(t *testing.T) {
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
				Green:  80,
				Yellow: 60,
				Red:    0,
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
				Value:     12,
				Timestamp: time.Now(),
			},
			{
				Reference: "test_cat.KRI.test_kri",
				Value:     3,
				Timestamp: time.Now(),
			},
		},
	}

	// Create a MetricsProcessor
	processor := NewMetricsProcessor(metricsConfig, leversConfig, metricsData)

	// Test GetMetricByReference
	metric, err := processor.GetMetricByReference("test_cat.KPI.test_kpi")
	if err != nil {
		t.Fatalf("Failed to get metric by reference: %v", err)
	}
	if metric.Value != 12 {
		t.Errorf("Expected metric value 12, got %f", metric.Value)
	}

	// Test GetMetricByReference with non-existent metric
	_, err = processor.GetMetricByReference("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent metric, got nil")
	}

	// Test UpdateMetric
	err = processor.UpdateMetric("test_cat.KPI.test_kpi", 15)
	if err != nil {
		t.Fatalf("Failed to update metric: %v", err)
	}

	updatedMetric, err := processor.GetMetricByReference("test_cat.KPI.test_kpi")
	if err != nil {
		t.Fatalf("Failed to get updated metric: %v", err)
	}
	if updatedMetric.Value != 15 {
		t.Errorf("Expected updated metric value 15, got %f", updatedMetric.Value)
	}

	// Test UpdateMetric with new metric
	err = processor.UpdateMetric("test_cat.KPI.new_metric", 42)
	if err != nil {
		t.Fatalf("Failed to add new metric: %v", err)
	}

	newMetric, err := processor.GetMetricByReference("test_cat.KPI.new_metric")
	if err != nil {
		t.Fatalf("Failed to get new metric: %v", err)
	}
	if newMetric.Value != 42 {
		t.Errorf("Expected new metric value 42, got %f", newMetric.Value)
	}

	// Test GetAllMetrics
	allMetrics := processor.GetAllMetrics()
	if len(allMetrics) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(allMetrics))
	}

	// Test GetMetricsByCategory
	categoryMetrics := processor.GetMetricsByCategory("test_cat")
	if len(categoryMetrics) != 3 {
		t.Errorf("Expected 3 metrics for category, got %d", len(categoryMetrics))
	}

	// Test GetCategoryByID
	category, err := processor.GetCategoryByID("test_cat")
	if err != nil {
		t.Fatalf("Failed to get category by ID: %v", err)
	}
	if category.Name != "Test Category" {
		t.Errorf("Expected category name 'Test Category', got '%s'", category.Name)
	}

	// Test GetAllCategories
	allCategories := processor.GetAllCategories()
	if len(allCategories) != 1 {
		t.Errorf("Expected 1 category, got %d", len(allCategories))
	}

	// Test isValidReference
	if !isValidReference("test_cat.KPI.test_kpi") {
		t.Error("Expected 'test_cat.KPI.test_kpi' to be a valid reference")
	}
	if isValidReference("invalid_reference") {
		t.Error("Expected 'invalid_reference' to be an invalid reference")
	}

	// Test GetMetricType
	metricType, err := GetMetricType("test_cat.KPI.test_kpi")
	if err != nil {
		t.Fatalf("Failed to get metric type: %v", err)
	}
	if metricType != "KPI" {
		t.Errorf("Expected metric type 'KPI', got '%s'", metricType)
	}

	// Test GetMetricDefinition
	metricDef, err := processor.GetMetricDefinition("test_cat.KPI.test_kpi")
	if err != nil {
		t.Fatalf("Failed to get metric definition: %v", err)
	}

	kpi, ok := metricDef.(KPI)
	if !ok {
		t.Fatal("Failed to cast metric definition to KPI")
	}
	if kpi.Name != "Test KPI" {
		t.Errorf("Expected KPI name 'Test KPI', got '%s'", kpi.Name)
	}
}
