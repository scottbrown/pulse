package pulse

import (
	"fmt"
	"strings"
)

// ScoreCalculator handles calculation of scores for metrics and categories
type ScoreCalculator struct {
	metricsProcessor *MetricsProcessor
}

// NewScoreCalculator creates a new ScoreCalculator
func NewScoreCalculator(metricsProcessor *MetricsProcessor) *ScoreCalculator {
	return &ScoreCalculator{
		metricsProcessor: metricsProcessor,
	}
}

// CalculateMetricScore calculates the score for a single metric
func (s *ScoreCalculator) CalculateMetricScore(metric Metric) (*MetricScore, error) {
	// Get the metric definition
	metricDef, err := s.metricsProcessor.GetMetricDefinition(metric.Reference)
	if err != nil {
		return nil, err
	}

	// Calculate score based on metric type
	var score int
	var status TrafficLightStatus

	metricType, err := GetMetricType(metric.Reference)
	if err != nil {
		return nil, err
	}

	if metricType == "KPI" {
		kpi, ok := metricDef.(KPI)
		if !ok {
			return nil, fmt.Errorf("failed to cast metric definition to KPI")
		}
		score = calculateKPIScore(metric.Value, kpi)
		status = determineStatus(score, s.metricsProcessor.leversConfig.Global.Thresholds)
	} else if metricType == "KRI" {
		kri, ok := metricDef.(KRI)
		if !ok {
			return nil, fmt.Errorf("failed to cast metric definition to KRI")
		}
		score = calculateKRIScore(metric.Value, kri)
		status = determineStatus(score, s.metricsProcessor.leversConfig.Global.Thresholds)
	} else {
		return nil, fmt.Errorf("unknown metric type: %s", metricType)
	}

	return &MetricScore{
		Reference: metric.Reference,
		Score:     score,
		Status:    status,
	}, nil
}

// calculateKPIScore calculates the score for a KPI metric
func calculateKPIScore(value float64, kpi KPI) int {
	// For KPIs, lower values might be better (e.g., remediation time)
	// or higher values might be better (e.g., compliance percentage)
	// This implementation assumes lower is better, adjust as needed

	if value <= float64(kpi.ScoringBands["band_5"]) {
		return 95 // Band 5: 90-100 points (using 95 as midpoint)
	} else if value <= float64(kpi.ScoringBands["band_4"]) {
		return 85 // Band 4: 80-89 points (using 85 as midpoint)
	} else if value <= float64(kpi.ScoringBands["band_3"]) {
		return 75 // Band 3: 70-79 points (using 75 as midpoint)
	} else if value <= float64(kpi.ScoringBands["band_2"]) {
		return 65 // Band 2: 60-69 points (using 65 as midpoint)
	} else {
		return 30 // Band 1: 0-59 points (using 30 as midpoint)
	}
}

// calculateKRIScore calculates the score for a KRI metric
func calculateKRIScore(value float64, kri KRI) int {
	// For KRIs, lower values are typically better (e.g., number of vulnerabilities)
	if value <= float64(kri.ScoringBands["band_5"]) {
		return 95 // Band 5: 90-100 points (using 95 as midpoint)
	} else if value <= float64(kri.ScoringBands["band_4"]) {
		return 85 // Band 4: 80-89 points (using 85 as midpoint)
	} else if value <= float64(kri.ScoringBands["band_3"]) {
		return 75 // Band 3: 70-79 points (using 75 as midpoint)
	} else if value <= float64(kri.ScoringBands["band_2"]) {
		return 65 // Band 2: 60-69 points (using 65 as midpoint)
	} else {
		return 30 // Band 1: 0-59 points (using 30 as midpoint)
	}
}

// determineStatus determines the traffic light status based on the score
func determineStatus(score int, thresholds Thresholds) TrafficLightStatus {
	if score >= thresholds.Green {
		return Green
	} else if score >= thresholds.Yellow {
		return Yellow
	} else {
		return Red
	}
}

// CalculateCategoryScore calculates the score for a category
func (s *ScoreCalculator) CalculateCategoryScore(categoryID string) (*CategoryScore, error) {
	// Get the category
	category, err := s.metricsProcessor.GetCategoryByID(categoryID)
	if err != nil {
		return nil, err
	}

	// Get metrics for this category
	categoryMetrics := s.metricsProcessor.GetMetricsByCategory(categoryID)
	if len(categoryMetrics) == 0 {
		return nil, fmt.Errorf("no metrics found for category: %s", categoryID)
	}

	// Calculate scores for each metric
	var metricScores []MetricScore
	var totalScore int
	for _, metric := range categoryMetrics {
		metricScore, err := s.CalculateMetricScore(metric)
		if err != nil {
			return nil, err
		}
		metricScores = append(metricScores, *metricScore)
		totalScore += metricScore.Score
	}

	// Calculate average score
	averageScore := totalScore / len(metricScores)

	// Determine status
	var status TrafficLightStatus

	// Check if there are category-specific thresholds
	if categoryThresholds, exists := s.metricsProcessor.leversConfig.Weights.CategoryThresholds[categoryID]; exists {
		status = determineStatus(averageScore, categoryThresholds)
	} else {
		status = determineStatus(averageScore, s.metricsProcessor.leversConfig.Global.Thresholds)
	}

	return &CategoryScore{
		ID:      categoryID,
		Name:    category.Name,
		Score:   averageScore,
		Status:  status,
		Metrics: metricScores,
	}, nil
}

// CalculateOverallScore calculates the overall security posture score
func (s *ScoreCalculator) CalculateOverallScore() (*OverallScore, error) {
	// Get all categories
	categories := s.metricsProcessor.GetAllCategories()
	if len(categories) == 0 {
		return nil, fmt.Errorf("no categories found")
	}

	// Calculate scores for each category
	var categoryScores []CategoryScore
	var weightedScoreSum float64
	var weightSum float64

	for _, category := range categories {
		categoryScore, err := s.CalculateCategoryScore(category.ID)
		if err != nil {
			// Skip categories with no metrics
			if strings.Contains(err.Error(), "no metrics found for category") {
				continue
			}
			return nil, err
		}

		categoryScores = append(categoryScores, *categoryScore)

		// Get weight for this category
		weight, exists := s.metricsProcessor.leversConfig.Weights.Categories[category.ID]
		if !exists {
			// Use equal weights if not specified
			weight = 1.0 / float64(len(categories))
		}

		weightedScoreSum += float64(categoryScore.Score) * weight
		weightSum += weight
	}

	if len(categoryScores) == 0 {
		return nil, fmt.Errorf("no categories with metrics found")
	}

	// Calculate weighted average score
	var overallScore int
	if weightSum > 0 {
		overallScore = int(weightedScoreSum / weightSum)
	} else {
		// Fallback to simple average if weights sum to 0
		var totalScore int
		for _, cs := range categoryScores {
			totalScore += cs.Score
		}
		overallScore = totalScore / len(categoryScores)
	}

	// Determine status
	status := determineStatus(overallScore, s.metricsProcessor.leversConfig.Global.Thresholds)

	return &OverallScore{
		Score:      overallScore,
		Status:     status,
		Categories: categoryScores,
	}, nil
}
