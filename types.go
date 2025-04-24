package pulse

import (
	"time"
)

// FloatPtr creates a pointer to a float64 value
// This is useful for creating min/max values for scoring bands
func FloatPtr(v float64) *float64 {
	return &v
}

// Category represents a security program category with KPIs and KRIs
type Category struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	KPIs        []KPI  `yaml:"kpis"`
	KRIs        []KRI  `yaml:"kris"`
}

// ScoringBand represents a single scoring band with min/max values and the resulting score
type ScoringBand struct {
	Min   *float64 `yaml:"min,omitempty"`
	Max   *float64 `yaml:"max,omitempty"`
	Score int      `yaml:"score"`
}

// KPI represents a Key Performance Indicator
type KPI struct {
	ID           string        `yaml:"id"`
	Name         string        `yaml:"name"`
	Description  string        `yaml:"description"`
	Unit         string        `yaml:"unit"`
	Target       float64       `yaml:"target"`
	ScoringBands []ScoringBand `yaml:"scoring_bands"`
}

// KRI represents a Key Risk Indicator
type KRI struct {
	ID           string        `yaml:"id"`
	Name         string        `yaml:"name"`
	Description  string        `yaml:"description"`
	Unit         string        `yaml:"unit"`
	Threshold    float64       `yaml:"threshold"`
	ScoringBands []ScoringBand `yaml:"scoring_bands"`
}

// MetricsConfig represents the structure of the metrics configuration file
type MetricsConfig struct {
	Categories []Category `yaml:"categories"`
}

// Metric represents a single metric measurement
type Metric struct {
	Reference  string    `yaml:"reference"`
	Value      float64   `yaml:"value"`
	Timestamp  time.Time `yaml:"timestamp"`
	SourceFile string    `yaml:"-"` // Source file for the metric (not stored in YAML)
}

// MetricsData represents the structure of the metrics data file
type MetricsData struct {
	Metrics []Metric `yaml:"metrics"`
}

// Thresholds represents the traffic light thresholds

// Thresholds represents the traffic light thresholds
type Thresholds struct {
	Green  int `yaml:"green"`  // Green threshold
	Yellow int `yaml:"yellow"` // Yellow threshold
	Red    int `yaml:"red"`    // Red threshold
}

// CategoryWeights represents the weights for each category
type CategoryWeights map[string]float64

// CategoryThresholds represents category-specific thresholds
type CategoryThresholds map[string]Thresholds

// Global represents global configuration settings
type Global struct {
	Thresholds Thresholds `yaml:"thresholds"`
}

// Weights represents the weights configuration
type Weights struct {
	Categories         CategoryWeights    `yaml:"categories"`
	CategoryThresholds CategoryThresholds `yaml:"category_thresholds"`
}

// LeversConfig represents the structure of the executive levers configuration file
type LeversConfig struct {
	Global  Global  `yaml:"global"`
	Weights Weights `yaml:"weights"`
}

// TrafficLightStatus represents the status in the traffic light model
type TrafficLightStatus string

const (
	Green  TrafficLightStatus = "green"
	Yellow TrafficLightStatus = "yellow"
	Red    TrafficLightStatus = "red"
)

// MetricScore represents a calculated score for a metric
type MetricScore struct {
	Reference string
	Score     int
	Status    TrafficLightStatus
}

// CategoryScore represents a calculated score for a category
type CategoryScore struct {
	ID      string
	Name    string
	Score   int
	Status  TrafficLightStatus
	Metrics []MetricScore
}

// OverallScore represents the overall security posture score
type OverallScore struct {
	Score      int
	Status     TrafficLightStatus
	Categories []CategoryScore
}
