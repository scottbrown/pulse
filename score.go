package pulse

import (
	"fmt"
	"sort"
	"strings"
)

// ScoringMethod defines the method used for score calculation
type ScoringMethod string

const (
	// MedianScoring uses median for category scores and weighted median for overall score
	MedianScoring ScoringMethod = "median"
	// AverageScoring uses average for category scores and weighted average for overall score
	AverageScoring ScoringMethod = "average"
)

// ScoreCalculator handles calculation of scores for metrics and categories
type ScoreCalculator struct {
	metricsProcessor *MetricsProcessor
	scoringMethod    ScoringMethod
}

// NewScoreCalculator creates a new ScoreCalculator
func NewScoreCalculator(metricsProcessor *MetricsProcessor, scoringMethod ScoringMethod) *ScoreCalculator {
	// Default to median scoring if not specified
	if scoringMethod == "" {
		scoringMethod = MedianScoring
	}

	return &ScoreCalculator{
		metricsProcessor: metricsProcessor,
		scoringMethod:    scoringMethod,
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
		status = determineStatus(score, s.metricsProcessor.leversConfig.Global.KPIThresholds)
	} else if metricType == "KRI" {
		kri, ok := metricDef.(KRI)
		if !ok {
			return nil, fmt.Errorf("failed to cast metric definition to KRI")
		}
		score = calculateKRIScore(metric.Value, kri)
		status = determineStatus(score, s.metricsProcessor.leversConfig.Global.KRIThresholds)
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
	// For KPIs, check each scoring band to find the appropriate score
	// Bands are expected to be ordered from highest score to lowest

	// If no scoring bands are defined, return a default score
	if len(kpi.ScoringBands) == 0 {
		return 50
	}

	// Check each scoring band
	for _, band := range kpi.ScoringBands {
		// Check if the value falls within this band's range
		inBand := true

		// Check minimum bound if it exists
		if band.Min != nil && value < *band.Min {
			inBand = false
		}

		// Check maximum bound if it exists
		if band.Max != nil && value > *band.Max {
			inBand = false
		}

		// If the value is in this band, return the score
		if inBand {
			return band.Score
		}
	}

	// If no band matches, return the lowest band's score or a default
	if len(kpi.ScoringBands) > 0 {
		return kpi.ScoringBands[len(kpi.ScoringBands)-1].Score
	}
	return 0
}

// calculateKRIScore calculates the score for a KRI metric
func calculateKRIScore(value float64, kri KRI) int {
	// For KRIs, use the same logic as KPIs
	// Bands are expected to be ordered from highest score to lowest

	// If no scoring bands are defined, return a default score
	if len(kri.ScoringBands) == 0 {
		return 50
	}

	// Check each scoring band
	for _, band := range kri.ScoringBands {
		// Check if the value falls within this band's range
		inBand := true

		// Check minimum bound if it exists
		if band.Min != nil && value < *band.Min {
			inBand = false
		}

		// Check maximum bound if it exists
		if band.Max != nil && value > *band.Max {
			inBand = false
		}

		// If the value is in this band, return the score
		if inBand {
			return band.Score
		}
	}

	// If no band matches, return the lowest band's score or a default
	if len(kri.ScoringBands) > 0 {
		return kri.ScoringBands[len(kri.ScoringBands)-1].Score
	}
	return 0
}

// determineStatus determines the traffic light status based on the score
func determineStatus(score int, thresholds Thresholds) TrafficLightStatus {
	// Check if score is within the Green range
	if score >= thresholds.Green.Min && score <= thresholds.Green.Max {
		return Green
	}

	// Check if score is within the Yellow range
	if score >= thresholds.Yellow.Min && score <= thresholds.Yellow.Max {
		return Yellow
	}

	// Default to Red
	return Red
}

// calculateAverage calculates the average value from a slice of integers
func calculateAverage(values []int) int {
	if len(values) == 0 {
		return 0
	}

	sum := 0
	for _, v := range values {
		sum += v
	}

	return sum / len(values)
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
	var kpiScores []int
	var kriScores []int

	for _, metric := range categoryMetrics {
		metricScore, err := s.CalculateMetricScore(metric)
		if err != nil {
			return nil, err
		}
		metricScores = append(metricScores, *metricScore)
		scores = append(scores, metricScore.Score)

		// Separate KPI and KRI scores
		metricType, err := GetMetricType(metric.Reference)
		if err != nil {
			return nil, err
		}

		if metricType == "KPI" {
			kpiScores = append(kpiScores, metricScore.Score)
		} else if metricType == "KRI" {
			kriScores = append(kriScores, metricScore.Score)
		}
	}

	// Calculate overall category score based on scoring method
	var categoryScore int
	if s.scoringMethod == MedianScoring {
		categoryScore = calculateMedian(scores)
	} else {
		categoryScore = calculateAverage(scores)
	}

	// Calculate separate KPI and KRI scores
	var kpiScore, kriScore int
	if len(kpiScores) > 0 {
		if s.scoringMethod == MedianScoring {
			kpiScore = calculateMedian(kpiScores)
		} else {
			kpiScore = calculateAverage(kpiScores)
		}
	}

	if len(kriScores) > 0 {
		if s.scoringMethod == MedianScoring {
			kriScore = calculateMedian(kriScores)
		} else {
			kriScore = calculateAverage(kriScores)
		}
	}

	// Determine overall status
	var status TrafficLightStatus

	// Check if there are category-specific thresholds
	if categoryThresholds, exists := s.metricsProcessor.leversConfig.Weights.CategoryThresholds[categoryID]; exists {
		status = determineStatus(categoryScore, categoryThresholds)
	} else {
		status = determineStatus(categoryScore, s.metricsProcessor.leversConfig.Global.Thresholds)
	}

	// Determine KPI status
	var kpiStatus TrafficLightStatus
	if len(kpiScores) > 0 {
		if categoryKPIThresholds, exists := s.metricsProcessor.leversConfig.Weights.CategoryKPIThresholds[categoryID]; exists {
			kpiStatus = determineStatus(kpiScore, categoryKPIThresholds)
		} else {
			kpiStatus = determineStatus(kpiScore, s.metricsProcessor.leversConfig.Global.KPIThresholds)
		}
	} else {
		kpiStatus = Yellow // Default if no KPIs
	}

	// Determine KRI status
	var kriStatus TrafficLightStatus
	if len(kriScores) > 0 {
		if categoryKRIThresholds, exists := s.metricsProcessor.leversConfig.Weights.CategoryKRIThresholds[categoryID]; exists {
			kriStatus = determineStatus(kriScore, categoryKRIThresholds)
		} else {
			kriStatus = determineStatus(kriScore, s.metricsProcessor.leversConfig.Global.KRIThresholds)
		}
	} else {
		kriStatus = Yellow // Default if no KRIs
	}

	return &CategoryScore{
		ID:        categoryID,
		Name:      category.Name,
		Score:     categoryScore,
		KPIScore:  kpiScore,
		KRIScore:  kriScore,
		Status:    status,
		KPIStatus: kpiStatus,
		KRIStatus: kriStatus,
		Metrics:   metricScores,
	}, nil
}

// calculateWeightedAverage calculates the weighted average from a slice of integers and their weights
func calculateWeightedAverage(values []int, weights []float64) int {
	if len(values) == 0 || len(values) != len(weights) {
		return 0
	}

	var weightedSum float64
	var totalWeight float64

	for i := 0; i < len(values); i++ {
		weightedSum += float64(values[i]) * weights[i]
		totalWeight += weights[i]
	}

	if totalWeight <= 0 {
		// If total weight is zero or negative, return simple average
		return calculateAverage(values)
	}

	return int(weightedSum / totalWeight)
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
	var kpiScores []int
	var kriScores []int
	var weights []float64
	var kpiWeights []float64
	var kriWeights []float64

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

		// Add KPI and KRI scores and weights if they exist
		if categoryScore.KPIScore > 0 {
			kpiScores = append(kpiScores, categoryScore.KPIScore)
			kpiWeights = append(kpiWeights, weight)
		}

		if categoryScore.KRIScore > 0 {
			kriScores = append(kriScores, categoryScore.KRIScore)
			kriWeights = append(kriWeights, weight)
		}
	}

	if len(categoryScores) == 0 {
		return nil, fmt.Errorf("no categories with metrics found")
	}

	// Calculate overall score as weighted sum of category scores
	var overallScore int
	var weightedSum float64
	for i := 0; i < len(scores); i++ {
		weightedSum += float64(scores[i]) * weights[i]
	}
	overallScore = int(weightedSum)

	// Calculate KPI and KRI scores
	var kpiScore, kriScore int

	if len(kpiScores) > 0 {
		// Always use weighted average for overall scores, regardless of scoring method
		kpiScore = calculateWeightedAverage(kpiScores, kpiWeights)
	}

	if len(kriScores) > 0 {
		// Always use weighted average for overall scores, regardless of scoring method
		kriScore = calculateWeightedAverage(kriScores, kriWeights)
	}

	// Determine overall status
	status := determineStatus(overallScore, s.metricsProcessor.leversConfig.Global.Thresholds)

	// Determine KPI status
	var kpiStatus TrafficLightStatus
	if len(kpiScores) > 0 {
		kpiStatus = determineStatus(kpiScore, s.metricsProcessor.leversConfig.Global.KPIThresholds)
	} else {
		kpiStatus = Yellow // Default if no KPIs
	}

	// Determine KRI status
	var kriStatus TrafficLightStatus
	if len(kriScores) > 0 {
		kriStatus = determineStatus(kriScore, s.metricsProcessor.leversConfig.Global.KRIThresholds)
	} else {
		kriStatus = Yellow // Default if no KRIs
	}

	return &OverallScore{
		Score:      overallScore,
		KPIScore:   kpiScore,
		KRIScore:   kriScore,
		Status:     status,
		KPIStatus:  kpiStatus,
		KRIStatus:  kriStatus,
		Categories: categoryScores,
	}, nil
}
