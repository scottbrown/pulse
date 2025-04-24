package pulse

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// MetricsProcessor handles processing and analysis of metrics
type MetricsProcessor struct {
	metricsConfig *MetricsConfig
	leversConfig  *LeversConfig
	metricsData   *MetricsData
}

// NewMetricsProcessor creates a new MetricsProcessor with the specified configurations
func NewMetricsProcessor(metricsConfig *MetricsConfig, leversConfig *LeversConfig, metricsData *MetricsData) *MetricsProcessor {
	return &MetricsProcessor{
		metricsConfig: metricsConfig,
		leversConfig:  leversConfig,
		metricsData:   metricsData,
	}
}

// GetMetricByReference returns a metric by its reference
func (m *MetricsProcessor) GetMetricByReference(reference string) (*Metric, error) {
	for _, metric := range m.metricsData.Metrics {
		if metric.Reference == reference {
			return &metric, nil
		}
	}
	return nil, fmt.Errorf("metric not found: %s", reference)
}

// UpdateMetric updates a metric value or adds a new metric if it doesn't exist
func (m *MetricsProcessor) UpdateMetric(reference string, value float64) error {
	// Validate the reference format
	if !isValidReference(reference) {
		return fmt.Errorf("invalid metric reference format: %s", reference)
	}

	// Extract category ID from reference for file organization
	parts := strings.Split(reference, ".")
	categoryID := parts[0]
	sourceFile := categoryID + ".yaml"

	// Check if the metric exists
	var found bool
	for i, metric := range m.metricsData.Metrics {
		if metric.Reference == reference {
			m.metricsData.Metrics[i].Value = value
			m.metricsData.Metrics[i].Timestamp = time.Now()

			// Update source file if not already set
			if m.metricsData.Metrics[i].SourceFile == "" {
				m.metricsData.Metrics[i].SourceFile = sourceFile
			}

			found = true
			break
		}
	}

	// If not found, add a new metric
	if !found {
		m.metricsData.Metrics = append(m.metricsData.Metrics, Metric{
			Reference:  reference,
			Value:      value,
			Timestamp:  time.Now(),
			SourceFile: sourceFile,
		})
	}

	return nil
}

// GetAllMetrics returns all metrics
func (m *MetricsProcessor) GetAllMetrics() []Metric {
	return m.metricsData.Metrics
}

// GetMetricsByCategory returns metrics for a specific category
func (m *MetricsProcessor) GetMetricsByCategory(categoryID string) []Metric {
	var categoryMetrics []Metric

	for _, metric := range m.metricsData.Metrics {
		parts := strings.Split(metric.Reference, ".")
		if len(parts) >= 1 && parts[0] == categoryID {
			categoryMetrics = append(categoryMetrics, metric)
		}
	}

	return categoryMetrics
}

// GetCategoryByID returns a category by its ID
func (m *MetricsProcessor) GetCategoryByID(categoryID string) (*Category, error) {
	for _, category := range m.metricsConfig.Categories {
		if category.ID == categoryID {
			return &category, nil
		}
	}
	return nil, fmt.Errorf("category not found: %s", categoryID)
}

// GetAllCategories returns all categories
func (m *MetricsProcessor) GetAllCategories() []Category {
	return m.metricsConfig.Categories
}

// isValidReference checks if a metric reference has the correct format
func isValidReference(reference string) bool {
	// Check for empty or overly long references
	if reference == "" || len(reference) > 100 {
		return false
	}

	// Check for invalid characters
	for _, char := range reference {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '.' && char != '_' && char != '-' {
			return false
		}
	}

	parts := strings.Split(reference, ".")
	if len(parts) != 3 {
		return false
	}

	// Check if each part is not empty
	for _, part := range parts {
		if part == "" {
			return false
		}
	}

	// Check if the second part is KPI or KRI
	if parts[1] != "KPI" && parts[1] != "KRI" {
		return false
	}

	return true
}

// GetMetricType returns whether a metric is a KPI or KRI
func GetMetricType(reference string) (string, error) {
	parts := strings.Split(reference, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid metric reference format: %s", reference)
	}

	if parts[1] != "KPI" && parts[1] != "KRI" {
		return "", fmt.Errorf("invalid metric type: %s", parts[1])
	}

	return parts[1], nil
}

// GetMetricDefinition returns the KPI or KRI definition for a metric
func (m *MetricsProcessor) GetMetricDefinition(reference string) (interface{}, error) {
	parts := strings.Split(reference, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid metric reference format: %s", reference)
	}

	categoryID := parts[0]
	metricType := parts[1]
	metricID := parts[2]

	category, err := m.GetCategoryByID(categoryID)
	if err != nil {
		return nil, err
	}

	if metricType == "KPI" {
		for _, kpi := range category.KPIs {
			if kpi.ID == metricID {
				return kpi, nil
			}
		}
		return nil, fmt.Errorf("KPI not found: %s", metricID)
	} else if metricType == "KRI" {
		for _, kri := range category.KRIs {
			if kri.ID == metricID {
				return kri, nil
			}
		}
		return nil, fmt.Errorf("KRI not found: %s", metricID)
	}

	return nil, fmt.Errorf("invalid metric type: %s", metricType)
}
