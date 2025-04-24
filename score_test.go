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
						ScoringBands: map[string]int{
							"band_5": 95,
							"band_4": 90,
							"band_3": 85,
							"band_2": 80,
							"band_1": 79,
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
				"test_cat":  0.6,
				"test_cat2": 0.4,
			},
			CategoryThresholds: CategoryThresholds{
				"test_cat2": Thresholds{
					Green:  85,
					Yellow: 70,
					Red:    0,
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

	// Create a ScoreCalculator
	calculator := NewScoreCalculator(processor)

	// Test CalculateMetricScore for KPI
	kpiMetric := Metric{
		Reference: "test_cat.KPI.test_kpi",
		Value:     3,
	}
	kpiScore, err := calculator.CalculateMetricScore(kpiMetric)
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
	kriScore, err := calculator.CalculateMetricScore(kriMetric)
	if err != nil {
		t.Fatalf("Failed to calculate KRI score: %v", err)
	}
	if kriScore.Score != 75 {
		t.Errorf("Expected KRI score 75, got %d", kriScore.Score)
	}
	if kriScore.Status != Yellow {
		t.Errorf("Expected KRI status Yellow, got %s", kriScore.Status)
	}

	// Test CalculateCategoryScore
	categoryScore, err := calculator.CalculateCategoryScore("test_cat")
	if err != nil {
		t.Fatalf("Failed to calculate category score: %v", err)
	}
	if categoryScore.Score != 85 { // Median of 95 and 75
		t.Errorf("Expected category score 85, got %d", categoryScore.Score)
	}
	if categoryScore.Status != Green {
		t.Errorf("Expected category status Green, got %s", categoryScore.Status)
	}
	if len(categoryScore.Metrics) != 2 {
		t.Errorf("Expected 2 metrics in category score, got %d", len(categoryScore.Metrics))
	}

	// Test CalculateCategoryScore with category-specific thresholds
	categoryScore2, err := calculator.CalculateCategoryScore("test_cat2")
	if err != nil {
		t.Fatalf("Failed to calculate category 2 score: %v", err)
	}
	if categoryScore2.Score != 95 {
		t.Errorf("Expected category 2 score 95, got %d", categoryScore2.Score)
	}
	if categoryScore2.Status != Green {
		t.Errorf("Expected category 2 status Green, got %s", categoryScore2.Status)
	}

	// Test CalculateOverallScore
	overallScore, err := calculator.CalculateOverallScore()
	if err != nil {
		t.Fatalf("Failed to calculate overall score: %v", err)
	}

	// Expected overall score: weighted median of 85 (weight 0.6) and 95 (weight 0.4) = 85
	// Since 0.6 > 0.5, the weighted median is 85
	if overallScore.Score != 85 {
		t.Errorf("Expected overall score 85, got %d", overallScore.Score)
	}
	if overallScore.Status != Green {
		t.Errorf("Expected overall status Green, got %s", overallScore.Status)
	}
	if len(overallScore.Categories) != 2 {
		t.Errorf("Expected 2 categories in overall score, got %d", len(overallScore.Categories))
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
		ScoringBands: map[string]int{
			"band_5": 5,
			"band_4": 10,
			"band_3": 15,
			"band_2": 20,
			"band_1": 21,
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
		ScoringBands: map[string]int{
			"band_5": 0,
			"band_4": 2,
			"band_3": 5,
			"band_2": 10,
			"band_1": 11,
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
