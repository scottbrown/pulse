package pulse

import (
	"time"
)

// Category represents a security program category with KPIs and KRIs
type Category struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	KPIs        []KPI  `yaml:"kpis"`
	KRIs        []KRI  `yaml:"kris"`
}

// KPI represents a Key Performance Indicator
type KPI struct {
	ID           string         `yaml:"id"`
	Name         string         `yaml:"name"`
	Description  string         `yaml:"description"`
	Unit         string         `yaml:"unit"`
	Target       float64        `yaml:"target"`
	ScoringBands map[string]int `yaml:"scoring_bands"`
}

// KRI represents a Key Risk Indicator
type KRI struct {
	ID           string         `yaml:"id"`
	Name         string         `yaml:"name"`
	Description  string         `yaml:"description"`
	Unit         string         `yaml:"unit"`
	Threshold    float64        `yaml:"threshold"`
	ScoringBands map[string]int `yaml:"scoring_bands"`
}

// MetricsConfig represents the structure of the metrics configuration file
type MetricsConfig struct {
	Categories []Category `yaml:"categories"`
}

// Metric represents a single metric measurement
type Metric struct {
	Reference string    `yaml:"reference"`
	Value     float64   `yaml:"value"`
	Timestamp time.Time `yaml:"timestamp"`
}

// MetricsData represents the structure of the metrics data file
type MetricsData struct {
	Metrics []Metric `yaml:"metrics"`
}

// ScoringBands represents the scoring bands for the traffic light model
type ScoringBands struct {
	Band5 int `yaml:"band_5"` // 90-100 points
	Band4 int `yaml:"band_4"` // 80-89 points
	Band3 int `yaml:"band_3"` // 70-79 points
	Band2 int `yaml:"band_2"` // 60-69 points
	Band1 int `yaml:"band_1"` // 0-59 points
}

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
	Thresholds   Thresholds   `yaml:"thresholds"`
	ScoringBands ScoringBands `yaml:"scoring_bands"`
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
