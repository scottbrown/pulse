package pulse

import (
	"testing"
	"time"
)

func TestScoreCalculator(t *testing.T) {
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
			{
				ID:          "test_cat2",
				Name:        "Test Category 2",
				Description: "Test category 2 description",
				KPIs: []KPI{
					{
						ID:          "test_kpi2",
						Name:        "Test KPI 2",
						Description: "Test KPI 2 description",
						Unit:        "percent",
						Target:      95,
						ScoringBands: []ScoringBand{
							{Min: FloatPtr(95), Score: 95},
							{Min: FloatPtr(90), Max: FloatPtr(95), Score: 85},
							{Min: FloatPtr(85), Max: FloatPtr(90), Score: 75},
							{Min: FloatPtr(80), Max: FloatPtr(85), Score: 65},
							{Max: FloatPtr(80), Score: 30},
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
				"test_cat":  0.6,
				"test_cat2": 0.4,
			},
			CategoryThresholds: CategoryThresholds{
				"test_cat2": Thresholds{
					Green: ThresholdRange{
						Min: 85,
						Max: 100,
					},
					Yellow: ThresholdRange{
						Min: 70,
						Max: 84,
					},
					Red: ThresholdRange{
						Min: 0,
						Max: 69,
					},
				},
			},
			CategoryKPIThresholds: CategoryThresholds{
				"test_cat2": Thresholds{
					Green: ThresholdRange{
						Min: 90,
						Max: 100,
					},
					Yellow: ThresholdRange{
						Min: 75,
						Max: 89,
					},
					Red: ThresholdRange{
						Min: 0,
						Max: 74,
					},
				},
			},
			CategoryKRIThresholds: CategoryThresholds{
				"test_cat2": Thresholds{
					Green: ThresholdRange{
						Min: 80,
						Max: 100,
					},
					Yellow: ThresholdRange{
						Min: 65,
						Max: 79,
					},
					Red: ThresholdRange{
						Min: 0,
						Max: 64,
					},
				},
			},
		},
	}

	metricsData := &MetricsData{
		Metrics: []Metric{
			{
				Reference: "test_cat.KPI.test_kpi",
				Value:     3, // Should be in band 5 (95 points)
				Timestamp: time.Now(),
			},
			{
				Reference: "test_cat.KRI.test_kri",
				Value:     4, // Should be in band 3 (75 points)
				Timestamp: time.Now(),
			},
			{
				Reference: "test_cat2.KPI.test_kpi2",
				Value:     92, // Should be in band 4 (85 points)
				Timestamp: time.Now(),
			},
		},
	}

	// Create a MetricsProcessor
	processor := NewMetricsProcessor(metricsConfig, leversConfig, metricsData)

	// Create ScoreCalculators for both scoring methods
	medianCalculator := NewScoreCalculator(processor, MedianScoring)
	averageCalculator := NewScoreCalculator(processor, AverageScoring)

	// Test CalculateMetricScore for KPI
	kpiMetric := Metric{
		Reference: "test_cat.KPI.test_kpi",
		Value:     3,
	}
	kpiScore, err := medianCalculator.CalculateMetricScore(kpiMetric)
	if err != nil {
		t.Fatalf("Failed to calculate KPI score: %v", err)
	}
	if kpiScore.Score != 95 {
		t.Errorf("Expected KPI score 95, got %d", kpiScore.Score)
	}
	if kpiScore.Status != Green {
		t.Errorf("Expected KPI status Green, got %s", kpiScore.Status)
	}

	// Test CalculateMetricScore for KRI
	kriMetric := Metric{
		Reference: "test_cat.KRI.test_kri",
		Value:     4,
	}
	kriScore, err := medianCalculator.CalculateMetricScore(kriMetric)
	if err != nil {
		t.Fatalf("Failed to calculate KRI score: %v", err)
	}
	if kriScore.Score != 75 {
		t.Errorf("Expected KRI score 75, got %d", kriScore.Score)
	}
	if kriScore.Status != Green {
		t.Errorf("Expected KRI status Green, got %s", kriScore.Status)
	}

	// Test CalculateCategoryScore with median scoring
	categoryScore, err := medianCalculator.CalculateCategoryScore("test_cat")
	if err != nil {
		t.Fatalf("Failed to calculate category score: %v", err)
	}
	if categoryScore.Score != 85 { // Median of 95 and 75
		t.Errorf("Expected category score 85, got %d", categoryScore.Score)
	}
	if categoryScore.KPIScore != 95 {
		t.Errorf("Expected category KPI score 95, got %d", categoryScore.KPIScore)
	}
	if categoryScore.KRIScore != 75 {
		t.Errorf("Expected category KRI score 75, got %d", categoryScore.KRIScore)
	}
	if categoryScore.Status != Green {
		t.Errorf("Expected category status Green, got %s", categoryScore.Status)
	}
	if categoryScore.KPIStatus != Green {
		t.Errorf("Expected category KPI status Green, got %s", categoryScore.KPIStatus)
	}
	if categoryScore.KRIStatus != Green {
		t.Errorf("Expected category KRI status Green, got %s", categoryScore.KRIStatus)
	}
	if len(categoryScore.Metrics) != 2 {
		t.Errorf("Expected 2 metrics in category score, got %d", len(categoryScore.Metrics))
	}

	// Test CalculateCategoryScore with category-specific thresholds using median scoring
	categoryScore2, err := medianCalculator.CalculateCategoryScore("test_cat2")
	if err != nil {
		t.Fatalf("Failed to calculate category 2 score: %v", err)
	}
	if categoryScore2.Score != 85 {
		t.Errorf("Expected category 2 score 85, got %d", categoryScore2.Score)
	}
	if categoryScore2.Status != Green {
		t.Errorf("Expected category 2 status Green, got %s", categoryScore2.Status)
	}

	// Test CalculateOverallScore with median scoring
	overallScore, err := medianCalculator.CalculateOverallScore()
	if err != nil {
		t.Fatalf("Failed to calculate overall score: %v", err)
	}

	// Expected overall score: weighted median of 85 (weight 0.6) and 95 (weight 0.4) = 85
	// Since 0.6 > 0.5, the weighted median is 85
	if overallScore.Score != 85 {
		t.Errorf("Expected overall score 85, got %d", overallScore.Score)
	}
	if overallScore.KPIScore != 91 {
		t.Errorf("Expected overall KPI score 91, got %d", overallScore.KPIScore)
	}
	if overallScore.KRIScore != 75 {
		t.Errorf("Expected overall KRI score 75, got %d", overallScore.KRIScore)
	}
	if overallScore.Status != Green {
		t.Errorf("Expected overall status Green, got %s", overallScore.Status)
	}
	if overallScore.KPIStatus != Green {
		t.Errorf("Expected overall KPI status Green, got %s", overallScore.KPIStatus)
	}
	if overallScore.KRIStatus != Green {
		t.Errorf("Expected overall KRI status Green, got %s", overallScore.KRIStatus)
	}
	if len(overallScore.Categories) != 2 {
		t.Errorf("Expected 2 categories in overall score, got %d", len(overallScore.Categories))
	}

	// Test CalculateCategoryScore with average scoring
	avgCategoryScore, err := averageCalculator.CalculateCategoryScore("test_cat")
	if err != nil {
		t.Fatalf("Failed to calculate category score with average scoring: %v", err)
	}
	if avgCategoryScore.Score != 85 { // Average of 95 and 75
		t.Errorf("Expected average category score 85, got %d", avgCategoryScore.Score)
	}

	// Test CalculateOverallScore with average scoring
	avgOverallScore, err := averageCalculator.CalculateOverallScore()
	if err != nil {
		t.Fatalf("Failed to calculate overall score with average scoring: %v", err)
	}

	// Expected overall score: (85 * 0.6) + (85 * 0.4) = 85
	if avgOverallScore.Score != 85 {
		t.Errorf("Expected average overall score 85, got %d", avgOverallScore.Score)
	}

	// Test determineStatus
	if determineStatus(85, leversConfig.Global.Thresholds) != Green {
		t.Error("Expected status Green for score 85")
	}
	if determineStatus(70, leversConfig.Global.Thresholds) != Yellow {
		t.Error("Expected status Yellow for score 70")
	}
	if determineStatus(50, leversConfig.Global.Thresholds) != Red {
		t.Error("Expected status Red for score 50")
	}

	// Test calculateKPIScore
	kpi := KPI{
		ScoringBands: []ScoringBand{
			{Max: FloatPtr(5), Score: 95},
			{Min: FloatPtr(5), Max: FloatPtr(10), Score: 85},
			{Min: FloatPtr(10), Max: FloatPtr(15), Score: 75},
			{Min: FloatPtr(15), Max: FloatPtr(20), Score: 65},
			{Min: FloatPtr(20), Score: 30},
		},
	}
	if calculateKPIScore(3, kpi) != 95 {
		t.Errorf("Expected KPI score 95 for value 3, got %d", calculateKPIScore(3, kpi))
	}
	if calculateKPIScore(12, kpi) != 75 {
		t.Errorf("Expected KPI score 75 for value 12, got %d", calculateKPIScore(12, kpi))
	}
	if calculateKPIScore(25, kpi) != 30 {
		t.Errorf("Expected KPI score 30 for value 25, got %d", calculateKPIScore(25, kpi))
	}

	// Test calculateKRIScore
	kri := KRI{
		ScoringBands: []ScoringBand{
			{Max: FloatPtr(0), Score: 95},
			{Min: FloatPtr(0), Max: FloatPtr(2), Score: 85},
			{Min: FloatPtr(2), Max: FloatPtr(5), Score: 75},
			{Min: FloatPtr(5), Max: FloatPtr(10), Score: 65},
			{Min: FloatPtr(10), Score: 30},
		},
	}
	if calculateKRIScore(0, kri) != 95 {
		t.Errorf("Expected KRI score 95 for value 0, got %d", calculateKRIScore(0, kri))
	}
	if calculateKRIScore(4, kri) != 75 {
		t.Errorf("Expected KRI score 75 for value 4, got %d", calculateKRIScore(4, kri))
	}
	if calculateKRIScore(15, kri) != 30 {
		t.Errorf("Expected KRI score 30 for value 15, got %d", calculateKRIScore(15, kri))
	}
}
