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

	"github.com/phpdave11/gofpdf"
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
	PDFFormat   ReportFormat = "pdf"
)

// ReportOutput represents the output of a report generation
type ReportOutput struct {
	Content     []byte
	ContentType string // "text" or "binary"
}

// GenerateOverallReport generates an overall security posture report
func (r *ReportGenerator) GenerateOverallReport(format ReportFormat) (*ReportOutput, error) {
	overallScore, err := r.scoreCalculator.CalculateOverallScore()
	if err != nil {
		return nil, err
	}

	switch format {
	case TextFormat:
		content := r.formatOverallReportAsText(overallScore)
		return &ReportOutput{Content: []byte(content), ContentType: "text"}, nil
	case JSONFormat:
		content, err := r.formatOverallReportAsJSON(overallScore)
		if err != nil {
			return nil, err
		}
		return &ReportOutput{Content: []byte(content), ContentType: "text"}, nil
	case TableFormat:
		content := r.formatOverallReportAsTable(overallScore)
		return &ReportOutput{Content: []byte(content), ContentType: "text"}, nil
	case PDFFormat:
		content, err := r.formatOverallReportAsPDF(overallScore)
		if err != nil {
			return nil, err
		}
		return &ReportOutput{Content: content, ContentType: "binary"}, nil
	default:
		return nil, fmt.Errorf("unsupported report format: %s", format)
	}
}

// GenerateCategoryReport generates a report for a specific category
func (r *ReportGenerator) GenerateCategoryReport(categoryID string, format ReportFormat) (*ReportOutput, error) {
	categoryScore, err := r.scoreCalculator.CalculateCategoryScore(categoryID)
	if err != nil {
		return nil, err
	}

	switch format {
	case TextFormat:
		content := r.formatCategoryReportAsText(categoryScore)
		return &ReportOutput{Content: []byte(content), ContentType: "text"}, nil
	case JSONFormat:
		content, err := r.formatCategoryReportAsJSON(categoryScore)
		if err != nil {
			return nil, err
		}
		return &ReportOutput{Content: []byte(content), ContentType: "text"}, nil
	case TableFormat:
		content := r.formatCategoryReportAsTable(categoryScore)
		return &ReportOutput{Content: []byte(content), ContentType: "text"}, nil
	case PDFFormat:
		content, err := r.formatCategoryReportAsPDF(categoryScore)
		if err != nil {
			return nil, err
		}
		return &ReportOutput{Content: content, ContentType: "binary"}, nil
	default:
		return nil, fmt.Errorf("unsupported report format: %s", format)
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

// formatOverallReportAsPDF formats the overall report as a PDF
func (r *ReportGenerator) formatOverallReportAsPDF(score *OverallScore) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set up fonts
	pdf.SetFont("Arial", "B", 16)

	// Title
	pdf.CellFormat(190, 10, "SECURITY POSTURE REPORT", "", 1, "C", false, 0, "")
	pdf.Ln(15)

	// Summary
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(40, 10, "KPI Score:", "", 0, "", false, 0, "")
	pdf.CellFormat(20, 10, fmt.Sprintf("%d", score.KPIScore), "", 0, "", false, 0, "")
	r.formatPDFStatus(pdf, score.KPIStatus)
	pdf.Ln(10)

	pdf.CellFormat(40, 10, "KRI Score:", "", 0, "", false, 0, "")
	pdf.CellFormat(20, 10, fmt.Sprintf("%d", score.KRIScore), "", 0, "", false, 0, "")
	r.formatPDFStatus(pdf, score.KRIStatus)
	pdf.Ln(10)

	pdf.CellFormat(40, 10, "Report Date:", "", 0, "", false, 0, "")
	pdf.CellFormat(60, 10, time.Now().Format("2006-01-02 15:04:05"), "", 0, "", false, 0, "")
	pdf.Ln(15)

	// Category scores table
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 10, "CATEGORY SCORES:", "", 1, "", false, 0, "")
	pdf.Ln(10)

	// Table header
	pdf.SetFillColor(200, 200, 200)
	pdf.SetFont("Arial", "B", 10)

	// Define table dimensions
	colWidths := []float64{60, 20, 25, 30, 25, 30}

	pdf.CellFormat(colWidths[0], 10, "Category", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths[1], 10, "Weight", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths[2], 10, "KPI Score", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths[3], 10, "KPI Status", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths[4], 10, "KRI Score", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths[5], 10, "KRI Status", "1", 1, "C", true, 0, "")

	// Table rows
	pdf.SetFont("Arial", "", 10)
	for _, category := range score.Categories {
		// Get the weight for this category
		weight, exists := r.scoreCalculator.metricsProcessor.leversConfig.Weights.Categories[category.ID]
		if !exists {
			// Use equal weights if not specified
			weight = 1.0 / float64(len(score.Categories))
		}

		// Format weight as percentage
		weightPercentage := int(weight * 100)

		// Draw the row cells
		pdf.CellFormat(colWidths[0], 10, sanitizeString(category.Name), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[1], 10, fmt.Sprintf("%d%%", weightPercentage), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[2], 10, fmt.Sprintf("%d", category.KPIScore), "1", 0, "C", false, 0, "")

		// KPI Status
		statusText := ""
		switch category.KPIStatus {
		case Green:
			pdf.SetTextColor(0, 128, 0) // Dark green
			statusText = "GREEN"
		case Yellow:
			pdf.SetTextColor(255, 165, 0) // Orange
			statusText = "YELLOW"
		case Red:
			pdf.SetTextColor(255, 0, 0) // Red
			statusText = "RED"
		default:
			pdf.SetTextColor(128, 128, 128) // Gray
			statusText = "UNKNOWN"
		}
		pdf.CellFormat(colWidths[3], 10, statusText, "1", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0) // Reset to black

		// KRI Score and Status
		pdf.CellFormat(colWidths[4], 10, fmt.Sprintf("%d", category.KRIScore), "1", 0, "C", false, 0, "")

		// KRI Status
		statusText = ""
		switch category.KRIStatus {
		case Green:
			pdf.SetTextColor(0, 128, 0) // Dark green
			statusText = "GREEN"
		case Yellow:
			pdf.SetTextColor(255, 165, 0) // Orange
			statusText = "YELLOW"
		case Red:
			pdf.SetTextColor(255, 0, 0) // Red
			statusText = "RED"
		default:
			pdf.SetTextColor(128, 128, 128) // Gray
			statusText = "UNKNOWN"
		}
		pdf.CellFormat(colWidths[5], 10, statusText, "1", 1, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0) // Reset to black
	}

	pdf.Ln(15)

	// Detailed metrics table
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 10, "DETAILED METRICS:", "", 1, "", false, 0, "")
	pdf.Ln(10)

	// Table header
	pdf.SetFillColor(200, 200, 200)
	pdf.SetFont("Arial", "B", 10)

	// Define table dimensions for detailed metrics
	detailColWidths := []float64{60, 30, 30, 30, 40}

	pdf.CellFormat(detailColWidths[0], 10, "Category", "1", 0, "C", true, 0, "")
	pdf.CellFormat(detailColWidths[1], 10, "Metric Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(detailColWidths[2], 10, "Metric ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(detailColWidths[3], 10, "Score", "1", 0, "C", true, 0, "")
	pdf.CellFormat(detailColWidths[4], 10, "Status", "1", 1, "C", true, 0, "")

	// Table rows
	pdf.SetFont("Arial", "", 10)
	for _, category := range score.Categories {
		for _, metric := range category.Metrics {
			parts := strings.Split(metric.Reference, ".")
			if len(parts) == 3 {
				metricType := parts[1]
				metricID := parts[2]

				// Draw the row cells
				pdf.CellFormat(detailColWidths[0], 10, sanitizeString(category.Name), "1", 0, "L", false, 0, "")
				pdf.CellFormat(detailColWidths[1], 10, sanitizeString(metricType), "1", 0, "C", false, 0, "")
				pdf.CellFormat(detailColWidths[2], 10, sanitizeString(metricID), "1", 0, "C", false, 0, "")
				pdf.CellFormat(detailColWidths[3], 10, fmt.Sprintf("%d", metric.Score), "1", 0, "C", false, 0, "")

				// Status
				statusText := ""
				switch metric.Status {
				case Green:
					pdf.SetTextColor(0, 128, 0) // Dark green
					statusText = "GREEN"
				case Yellow:
					pdf.SetTextColor(255, 165, 0) // Orange
					statusText = "YELLOW"
				case Red:
					pdf.SetTextColor(255, 0, 0) // Red
					statusText = "RED"
				default:
					pdf.SetTextColor(128, 128, 128) // Gray
					statusText = "UNKNOWN"
				}
				pdf.CellFormat(detailColWidths[4], 10, statusText, "1", 1, "C", false, 0, "")
				pdf.SetTextColor(0, 0, 0) // Reset to black
			}
		}
	}

	// Add page numbers
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()), "0", 0, "C", false, 0, "")
	})

	// Output the PDF as bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// formatCategoryReportAsPDF formats a category report as a PDF
func (r *ReportGenerator) formatCategoryReportAsPDF(score *CategoryScore) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

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

	// Set up fonts
	pdf.SetFont("Arial", "B", 16)

	// Title
	pdf.CellFormat(190, 10, fmt.Sprintf("%s REPORT (WEIGHT: %d%%)", strings.ToUpper(sanitizeString(score.Name)), weightPercentage), "", 1, "", false, 0, "")
	pdf.Ln(15)

	// Summary
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(40, 10, "KPI Score:", "", 0, "", false, 0, "")
	pdf.CellFormat(20, 10, fmt.Sprintf("%d", score.KPIScore), "", 0, "", false, 0, "")
	r.formatPDFStatus(pdf, score.KPIStatus)
	pdf.Ln(10)

	pdf.CellFormat(40, 10, "KRI Score:", "", 0, "", false, 0, "")
	pdf.CellFormat(20, 10, fmt.Sprintf("%d", score.KRIScore), "", 0, "", false, 0, "")
	r.formatPDFStatus(pdf, score.KRIStatus)
	pdf.Ln(10)

	pdf.CellFormat(40, 10, "Report Date:", "", 0, "", false, 0, "")
	pdf.CellFormat(60, 10, time.Now().Format("2006-01-02 15:04:05"), "", 0, "", false, 0, "")
	pdf.Ln(15)

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

	// Metrics table
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 10, "METRICS:", "", 1, "", false, 0, "")
	pdf.Ln(10)

	// Table header
	pdf.SetFillColor(200, 200, 200)
	pdf.SetFont("Arial", "B", 10)

	// Define table dimensions
	colWidths := []float64{30, 60, 30, 70}

	pdf.CellFormat(colWidths[0], 10, "Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths[1], 10, "ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths[2], 10, "Score", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths[3], 10, "Status", "1", 1, "C", true, 0, "")

	// Table rows for KPIs
	pdf.SetFont("Arial", "", 10)
	for _, metric := range kpiMetrics {
		parts := strings.Split(metric.Reference, ".")
		if len(parts) == 3 {
			metricID := parts[2]

			// Draw the row cells
			pdf.CellFormat(colWidths[0], 10, "KPI", "1", 0, "C", false, 0, "")
			pdf.CellFormat(colWidths[1], 10, sanitizeString(metricID), "1", 0, "L", false, 0, "")
			pdf.CellFormat(colWidths[2], 10, fmt.Sprintf("%d", metric.Score), "1", 0, "C", false, 0, "")

			// Status
			statusText := ""
			switch metric.Status {
			case Green:
				pdf.SetTextColor(0, 128, 0) // Dark green
				statusText = "GREEN"
			case Yellow:
				pdf.SetTextColor(255, 165, 0) // Orange
				statusText = "YELLOW"
			case Red:
				pdf.SetTextColor(255, 0, 0) // Red
				statusText = "RED"
			default:
				pdf.SetTextColor(128, 128, 128) // Gray
				statusText = "UNKNOWN"
			}
			pdf.CellFormat(colWidths[3], 10, statusText, "1", 1, "C", false, 0, "")
			pdf.SetTextColor(0, 0, 0) // Reset to black
		}
	}

	// Table rows for KRIs
	for _, metric := range kriMetrics {
		parts := strings.Split(metric.Reference, ".")
		if len(parts) == 3 {
			metricID := parts[2]

			// Draw the row cells
			pdf.CellFormat(colWidths[0], 10, "KRI", "1", 0, "C", false, 0, "")
			pdf.CellFormat(colWidths[1], 10, sanitizeString(metricID), "1", 0, "L", false, 0, "")
			pdf.CellFormat(colWidths[2], 10, fmt.Sprintf("%d", metric.Score), "1", 0, "C", false, 0, "")

			// Status
			statusText := ""
			switch metric.Status {
			case Green:
				pdf.SetTextColor(0, 128, 0) // Dark green
				statusText = "GREEN"
			case Yellow:
				pdf.SetTextColor(255, 165, 0) // Orange
				statusText = "YELLOW"
			case Red:
				pdf.SetTextColor(255, 0, 0) // Red
				statusText = "RED"
			default:
				pdf.SetTextColor(128, 128, 128) // Gray
				statusText = "UNKNOWN"
			}
			pdf.CellFormat(colWidths[3], 10, statusText, "1", 1, "C", false, 0, "")
			pdf.SetTextColor(0, 0, 0) // Reset to black
		}
	}

	// Add page numbers
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()), "0", 0, "C", false, 0, "")
	})

	// Output the PDF as bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// formatPDFStatus formats a traffic light status for display in PDF
func (r *ReportGenerator) formatPDFStatus(pdf *gofpdf.Fpdf, status TrafficLightStatus) string {
	switch status {
	case Green:
		pdf.SetTextColor(0, 128, 0) // Dark green
		// Always use text labels for PDF to avoid encoding issues
		pdf.CellFormat(30, 10, "GREEN", "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0) // Reset to black
		return ""
	case Yellow:
		pdf.SetTextColor(255, 165, 0) // Orange
		pdf.CellFormat(30, 10, "YELLOW", "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0) // Reset to black
		return ""
	case Red:
		pdf.SetTextColor(255, 0, 0) // Red
		pdf.CellFormat(30, 10, "RED", "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0) // Reset to black
		return ""
	default:
		pdf.SetTextColor(128, 128, 128) // Gray
		pdf.CellFormat(30, 10, "UNKNOWN", "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0) // Reset to black
		return ""
	}
}
