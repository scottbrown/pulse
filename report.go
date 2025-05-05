package pulse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"
)

// JSON report types
type jsonReport struct {
	ReportDate string         `json:"report_date"`
	KPIScore   int            `json:"kpi_score"`
	KRIScore   int            `json:"kri_score"`
	KPIStatus  string         `json:"kpi_status"`
	KRIStatus  string         `json:"kri_status"`
	Categories []jsonCategory `json:"categories"`
}

type jsonCategory struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	WeightPercent int          `json:"weight_percent"`
	KPIScore      int          `json:"kpi_score"`
	KRIScore      int          `json:"kri_score"`
	KPIStatus     string       `json:"kpi_status"`
	KRIStatus     string       `json:"kri_status"`
	Metrics       []jsonMetric `json:"metrics"`
}

type jsonMetric struct {
	Reference string `json:"reference"`
	Score     int    `json:"score"`
	Status    string `json:"status"`
}

type jsonCategoryReport struct {
	ReportDate    string       `json:"report_date"`
	CategoryID    string       `json:"category_id"`
	CategoryName  string       `json:"category_name"`
	WeightPercent int          `json:"weight_percent"`
	KPIScore      int          `json:"kpi_score"`
	KRIScore      int          `json:"kri_score"`
	KPIStatus     string       `json:"kpi_status"`
	KRIStatus     string       `json:"kri_status"`
	Metrics       []jsonMetric `json:"metrics"`
}

// ThresholdLabelType defines the type of threshold labels to use
type ThresholdLabelType string

const (
	// EmojiLabels uses emoji symbols for threshold labels
	EmojiLabels ThresholdLabelType = "emoji"
	// TextLabels uses text labels for thresholds
	TextLabels ThresholdLabelType = "text"
)

// ReportGenerator handles generation of reports
type ReportGenerator struct {
	scoreCalculator *ScoreCalculator
	labelType       ThresholdLabelType
}

// NewReportGenerator creates a new ReportGenerator
func NewReportGenerator(scoreCalculator *ScoreCalculator, labelType ThresholdLabelType) *ReportGenerator {
	return &ReportGenerator{
		scoreCalculator: scoreCalculator,
		labelType:       labelType,
	}
}

// ReportFormat defines the format of the report
type ReportFormat string

const (
	TextFormat  ReportFormat = "text"
	JSONFormat  ReportFormat = "json"
	TableFormat ReportFormat = "table"
)

// GenerateOverallReport generates an overall security posture report
func (r *ReportGenerator) GenerateOverallReport(format ReportFormat) (string, error) {
	overallScore, err := r.scoreCalculator.CalculateOverallScore()
	if err != nil {
		return "", err
	}

	switch format {
	case TextFormat:
		return r.formatOverallReportAsText(overallScore), nil
	case JSONFormat:
		return r.formatOverallReportAsJSON(overallScore)
	case TableFormat:
		return r.formatOverallReportAsTable(overallScore), nil
	default:
		return "", fmt.Errorf("unsupported report format: %s", format)
	}
}

// GenerateCategoryReport generates a report for a specific category
func (r *ReportGenerator) GenerateCategoryReport(categoryID string, format ReportFormat) (string, error) {
	categoryScore, err := r.scoreCalculator.CalculateCategoryScore(categoryID)
	if err != nil {
		return "", err
	}

	switch format {
	case TextFormat:
		return r.formatCategoryReportAsText(categoryScore), nil
	case JSONFormat:
		return r.formatCategoryReportAsJSON(categoryScore)
	case TableFormat:
		return r.formatCategoryReportAsTable(categoryScore), nil
	default:
		return "", fmt.Errorf("unsupported report format: %s", format)
	}
}

// sanitizeString sanitizes a string for safe output
func sanitizeString(input string) string {
	// Remove any control characters
	re := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	sanitized := re.ReplaceAllString(input, "")

	// Escape HTML entities for additional safety
	sanitized = html.EscapeString(sanitized)

	// Limit string length
	if len(sanitized) > 1000 {
		sanitized = sanitized[:1000] + "..."
	}

	return sanitized
}

// formatOverallReportAsText formats the overall report as text
func (r *ReportGenerator) formatOverallReportAsText(score *OverallScore) string {
	var sb strings.Builder

	sb.WriteString("===== SECURITY POSTURE REPORT =====\n\n")
	sb.WriteString(fmt.Sprintf("KPI Score: %d (%s)\n", score.KPIScore, r.formatStatus(score.KPIStatus)))
	sb.WriteString(fmt.Sprintf("KRI Score: %d (%s)\n", score.KRIScore, r.formatStatus(score.KRIStatus)))
	sb.WriteString(fmt.Sprintf("Report Date: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	sb.WriteString("Category Scores:\n")
	sb.WriteString("----------------\n")
	for _, category := range score.Categories {
		// Get the weight for this category
		weight, exists := r.scoreCalculator.metricsProcessor.leversConfig.Weights.Categories[category.ID]
		if !exists {
			// Use equal weights if not specified
			weight = 1.0 / float64(len(score.Categories))
		}

		// Format weight as percentage
		weightPercentage := int(weight * 100)

		sb.WriteString(fmt.Sprintf("- %s (weight: %d%%):\n", sanitizeString(category.Name), weightPercentage))
		sb.WriteString(fmt.Sprintf("  KPI: %d (%s), KRI: %d (%s)\n",
			category.KPIScore, r.formatStatus(category.KPIStatus),
			category.KRIScore, r.formatStatus(category.KRIStatus)))
	}

	sb.WriteString("\nDetailed Metrics:\n")
	sb.WriteString("----------------\n")
	for _, category := range score.Categories {
		sb.WriteString(fmt.Sprintf("\n%s:\n", sanitizeString(category.Name)))
		for _, metric := range category.Metrics {
			parts := strings.Split(metric.Reference, ".")
			if len(parts) == 3 {
				metricType := parts[1]
				metricID := parts[2]
				sb.WriteString(fmt.Sprintf("  - %s %s: %d (%s)\n", sanitizeString(metricType), sanitizeString(metricID), metric.Score, r.formatStatus(metric.Status)))
			}
		}
	}

	return sb.String()
}

// formatCategoryReportAsText formats a category report as text
func (r *ReportGenerator) formatCategoryReportAsText(score *CategoryScore) string {
	var sb strings.Builder

	// Get the weight for this category
	weight, exists := r.scoreCalculator.metricsProcessor.leversConfig.Weights.Categories[score.ID]
	if !exists {
		// Use equal weights if not specified
		totalCategories := len(r.scoreCalculator.metricsProcessor.GetAllCategories())
		if totalCategories > 0 {
			weight = 1.0 / float64(totalCategories)
		} else {
			weight = 1.0
		}
	}

	// Format weight as percentage
	weightPercentage := int(weight * 100)

	sb.WriteString(fmt.Sprintf("===== %s REPORT (WEIGHT: %d%%) =====\n\n", strings.ToUpper(sanitizeString(score.Name)), weightPercentage))
	sb.WriteString(fmt.Sprintf("KPI Score: %d (%s)\n", score.KPIScore, r.formatStatus(score.KPIStatus)))
	sb.WriteString(fmt.Sprintf("KRI Score: %d (%s)\n", score.KRIScore, r.formatStatus(score.KRIStatus)))
	sb.WriteString(fmt.Sprintf("Report Date: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	sb.WriteString("Metrics:\n")
	sb.WriteString("--------\n")

	// Group metrics by type
	var kpiMetrics []MetricScore
	var kriMetrics []MetricScore

	for _, metric := range score.Metrics {
		parts := strings.Split(metric.Reference, ".")
		if len(parts) == 3 {
			metricType := parts[1]
			if metricType == "KPI" {
				kpiMetrics = append(kpiMetrics, metric)
			} else if metricType == "KRI" {
				kriMetrics = append(kriMetrics, metric)
			}
		}
	}

	// Display KPIs
	if len(kpiMetrics) > 0 {
		sb.WriteString("\nKPIs:\n")
		for _, metric := range kpiMetrics {
			parts := strings.Split(metric.Reference, ".")
			if len(parts) == 3 {
				metricID := parts[2]
				sb.WriteString(fmt.Sprintf("- KPI %s: %d (%s)\n", sanitizeString(metricID), metric.Score, r.formatStatus(metric.Status)))
			}
		}
	}

	// Display KRIs
	if len(kriMetrics) > 0 {
		sb.WriteString("\nKRIs:\n")
		for _, metric := range kriMetrics {
			parts := strings.Split(metric.Reference, ".")
			if len(parts) == 3 {
				metricID := parts[2]
				sb.WriteString(fmt.Sprintf("- KRI %s: %d (%s)\n", sanitizeString(metricID), metric.Score, r.formatStatus(metric.Status)))
			}
		}
	}

	return sb.String()
}

// formatOverallReportAsJSON formats the overall report as JSON
func (r *ReportGenerator) formatOverallReportAsJSON(score *OverallScore) (string, error) {

	var categories []jsonCategory
	for _, category := range score.Categories {
		var metrics []jsonMetric
		for _, metric := range category.Metrics {
			metrics = append(metrics, jsonMetric{
				Reference: sanitizeString(metric.Reference),
				Score:     metric.Score,
				Status:    string(metric.Status),
			})
		}

		// Get the weight for this category
		weight, exists := r.scoreCalculator.metricsProcessor.leversConfig.Weights.Categories[category.ID]
		if !exists {
			// Use equal weights if not specified
			weight = 1.0 / float64(len(score.Categories))
		}

		// Format weight as percentage
		weightPercentage := int(weight * 100)

		categories = append(categories, jsonCategory{
			ID:            sanitizeString(category.ID),
			Name:          sanitizeString(category.Name),
			WeightPercent: weightPercentage,
			KPIScore:      category.KPIScore,
			KRIScore:      category.KRIScore,
			KPIStatus:     string(category.KPIStatus),
			KRIStatus:     string(category.KRIStatus),
			Metrics:       metrics,
		})
	}

	report := jsonReport{
		ReportDate: time.Now().Format(time.RFC3339),
		KPIScore:   score.KPIScore,
		KRIScore:   score.KRIScore,
		KPIStatus:  string(score.KPIStatus),
		KRIStatus:  string(score.KRIStatus),
		Categories: categories,
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	return string(jsonData), nil
}

// formatCategoryReportAsJSON formats a category report as JSON
func (r *ReportGenerator) formatCategoryReportAsJSON(score *CategoryScore) (string, error) {

	var metrics []jsonMetric
	for _, metric := range score.Metrics {
		metrics = append(metrics, jsonMetric{
			Reference: sanitizeString(metric.Reference),
			Score:     metric.Score,
			Status:    string(metric.Status),
		})
	}

	// Get the weight for this category
	weight, exists := r.scoreCalculator.metricsProcessor.leversConfig.Weights.Categories[score.ID]
	if !exists {
		// Use equal weights if not specified
		totalCategories := len(r.scoreCalculator.metricsProcessor.GetAllCategories())
		if totalCategories > 0 {
			weight = 1.0 / float64(totalCategories)
		} else {
			weight = 1.0
		}
	}

	// Format weight as percentage
	weightPercentage := int(weight * 100)

	report := jsonCategoryReport{
		ReportDate:    time.Now().Format(time.RFC3339),
		CategoryID:    sanitizeString(score.ID),
		CategoryName:  sanitizeString(score.Name),
		WeightPercent: weightPercentage,
		KPIScore:      score.KPIScore,
		KRIScore:      score.KRIScore,
		KPIStatus:     string(score.KPIStatus),
		KRIStatus:     string(score.KRIStatus),
		Metrics:       metrics,
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	return string(jsonData), nil
}

// formatStatus formats a traffic light status for display
func (r *ReportGenerator) formatStatus(status TrafficLightStatus) string {
	if r.labelType == TextLabels {
		switch status {
		case Green:
			return "GREEN"
		case Yellow:
			return "YELLOW"
		case Red:
			return "RED"
		default:
			return "UNKNOWN"
		}
	} else {
		// Default to emoji labels
		switch status {
		case Green:
			return "🟢" // Green circle
		case Yellow:
			return "🟡" // Yellow circle
		case Red:
			return "🔴" // Red circle
		default:
			return "❓" // Question mark
		}
	}
}

// formatOverallReportAsTable formats the overall report as a table
func (r *ReportGenerator) formatOverallReportAsTable(score *OverallScore) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	// Report header
	fmt.Fprintln(w, "===== SECURITY POSTURE REPORT =====")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "KPI Score:\t%d\t(%s)\n", score.KPIScore, r.formatStatus(score.KPIStatus))
	fmt.Fprintf(w, "KRI Score:\t%d\t(%s)\n", score.KRIScore, r.formatStatus(score.KRIStatus))
	fmt.Fprintf(w, "Report Date:\t%s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(w)

	// Category scores table
	fmt.Fprintln(w, "CATEGORY SCORES:")
	fmt.Fprintln(w, "Category\tWeight\tKPI Score\tKPI Status\tKRI Score\tKRI Status")
	fmt.Fprintln(w, "--------\t------\t---------\t----------\t---------\t----------")

	for _, category := range score.Categories {
		// Get the weight for this category
		weight, exists := r.scoreCalculator.metricsProcessor.leversConfig.Weights.Categories[category.ID]
		if !exists {
			// Use equal weights if not specified
			weight = 1.0 / float64(len(score.Categories))
		}

		// Format weight as percentage
		weightPercentage := int(weight * 100)

		fmt.Fprintf(w, "%s\t%d%%\t%d\t%s\t%d\t%s\n",
			sanitizeString(category.Name),
			weightPercentage,
			category.KPIScore,
			r.formatStatus(category.KPIStatus),
			category.KRIScore,
			r.formatStatus(category.KRIStatus))
	}
	fmt.Fprintln(w)

	// Detailed metrics table
	fmt.Fprintln(w, "DETAILED METRICS:")
	fmt.Fprintln(w, "Category\tMetric Type\tMetric ID\tScore\tStatus")
	fmt.Fprintln(w, "--------\t-----------\t---------\t-----\t------")

	for _, category := range score.Categories {
		for _, metric := range category.Metrics {
			parts := strings.Split(metric.Reference, ".")
			if len(parts) == 3 {
				metricType := parts[1]
				metricID := parts[2]
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
					sanitizeString(category.Name),
					sanitizeString(metricType),
					sanitizeString(metricID),
					metric.Score,
					r.formatStatus(metric.Status))
			}
		}
	}

	w.Flush()
	return buf.String()
}

// formatCategoryReportAsTable formats a category report as a table
func (r *ReportGenerator) formatCategoryReportAsTable(score *CategoryScore) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	// Get the weight for this category
	weight, exists := r.scoreCalculator.metricsProcessor.leversConfig.Weights.Categories[score.ID]
	if !exists {
		// Use equal weights if not specified
		totalCategories := len(r.scoreCalculator.metricsProcessor.GetAllCategories())
		if totalCategories > 0 {
			weight = 1.0 / float64(totalCategories)
		} else {
			weight = 1.0
		}
	}

	// Format weight as percentage
	weightPercentage := int(weight * 100)

	// Report header
	fmt.Fprintf(w, "===== %s REPORT (WEIGHT: %d%%) =====\n", strings.ToUpper(sanitizeString(score.Name)), weightPercentage)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "KPI Score:\t%d\t(%s)\n", score.KPIScore, r.formatStatus(score.KPIStatus))
	fmt.Fprintf(w, "KRI Score:\t%d\t(%s)\n", score.KRIScore, r.formatStatus(score.KRIStatus))
	fmt.Fprintf(w, "Report Date:\t%s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(w)

	// Group metrics by type
	var kpiMetrics []MetricScore
	var kriMetrics []MetricScore

	for _, metric := range score.Metrics {
		parts := strings.Split(metric.Reference, ".")
		if len(parts) == 3 {
			metricType := parts[1]
			if metricType == "KPI" {
				kpiMetrics = append(kpiMetrics, metric)
			} else if metricType == "KRI" {
				kriMetrics = append(kriMetrics, metric)
			}
		}
	}

	// Display metrics table
	fmt.Fprintln(w, "METRICS:")
	fmt.Fprintln(w, "Type\tID\tScore\tStatus")
	fmt.Fprintln(w, "----\t--\t-----\t------")

	// Display KPIs
	for _, metric := range kpiMetrics {
		parts := strings.Split(metric.Reference, ".")
		if len(parts) == 3 {
			metricID := parts[2]
			fmt.Fprintf(w, "KPI\t%s\t%d\t%s\n",
				sanitizeString(metricID),
				metric.Score,
				r.formatStatus(metric.Status))
		}
	}

	// Display KRIs
	for _, metric := range kriMetrics {
		parts := strings.Split(metric.Reference, ".")
		if len(parts) == 3 {
			metricID := parts[2]
			fmt.Fprintf(w, "KRI\t%s\t%d\t%s\n",
				sanitizeString(metricID),
				metric.Score,
				r.formatStatus(metric.Status))
		}
	}

	w.Flush()
	return buf.String()
}
