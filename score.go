package pulse

import (
	"fmt"
	"sort"
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

// calculateMedian calculates the median value from a slice of integers
func calculateMedian(values []int) int {
	if len(values) == 0 {
		return 0
	}

	// Create a copy of the slice to avoid modifying the original
	valuesCopy := make([]int, len(values))
	copy(valuesCopy, values)

	// Sort the values
	sort.Ints(valuesCopy)

	// Find the median
	middle := len(valuesCopy) / 2
	if len(valuesCopy)%2 == 0 {
		// Even number of elements, average the two middle values
		return (valuesCopy[middle-1] + valuesCopy[middle]) / 2
	}
	// Odd number of elements, return the middle value
	return valuesCopy[middle]
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
	var scores []int
	for _, metric := range categoryMetrics {
		metricScore, err := s.CalculateMetricScore(metric)
		if err != nil {
			return nil, err
		}
		metricScores = append(metricScores, *metricScore)
		scores = append(scores, metricScore.Score)
	}

	// Calculate median score
	medianScore := calculateMedian(scores)

	// Determine status
	var status TrafficLightStatus

	// Check if there are category-specific thresholds
	if categoryThresholds, exists := s.metricsProcessor.leversConfig.Weights.CategoryThresholds[categoryID]; exists {
		status = determineStatus(medianScore, categoryThresholds)
	} else {
		status = determineStatus(medianScore, s.metricsProcessor.leversConfig.Global.Thresholds)
	}

	return &CategoryScore{
		ID:      categoryID,
		Name:    category.Name,
		Score:   medianScore,
		Status:  status,
		Metrics: metricScores,
	}, nil
}

// calculateWeightedMedian calculates the weighted median from a slice of integers and their weights
func calculateWeightedMedian(values []int, weights []float64) int {
	if len(values) == 0 || len(values) != len(weights) {
		return 0
	}

	// Create pairs of values and weights
	type weightedValue struct {
		value  int
		weight float64
	}

	pairs := make([]weightedValue, len(values))
	for i := 0; i < len(values); i++ {
		pairs[i] = weightedValue{value: values[i], weight: weights[i]}
	}

	// Sort by value
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value < pairs[j].value
	})

	// Calculate total weight
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	if totalWeight <= 0 {
		// If total weight is zero or negative, return simple median
		return calculateMedian(values)
	}

	// Find the weighted median
	halfWeight := totalWeight / 2
	cumulativeWeight := 0.0

	for i, pair := range pairs {
		cumulativeWeight += pair.weight

		if cumulativeWeight > halfWeight {
			// Found the weighted median
			return pair.value
		} else if cumulativeWeight == halfWeight && i < len(pairs)-1 {
			// If we're exactly at half the weight and not at the last element,
			// the weighted median is the average of this value and the next
			return (pair.value + pairs[i+1].value) / 2
		}
	}

	// If we get here, return the last value
	return pairs[len(pairs)-1].value
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
	var scores []int
	var weights []float64

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

		scores = append(scores, categoryScore.Score)
		weights = append(weights, weight)
	}

	if len(categoryScores) == 0 {
		return nil, fmt.Errorf("no categories with metrics found")
	}

	// Calculate weighted median score
	overallScore := calculateWeightedMedian(scores, weights)

	// Determine status
	status := determineStatus(overallScore, s.metricsProcessor.leversConfig.Global.Thresholds)

	return &OverallScore{
		Score:      overallScore,
		Status:     status,
		Categories: categoryScores,
	}, nil
}
