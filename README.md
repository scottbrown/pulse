![Pulse](logo.small.png)

# Pulse

A Risk and Performance measurement framework CLI application for organizational programs (e.g.  Information Security, Legal).

## Overview

Pulse is a command-line tool designed to report on Key Performance Indicators (KPIs) and Key Risk Indicators (KRIs) for programs. It provides a flexible framework for defining, measuring, and reporting on metrics across multiple categories.

## Features

- **YAML-based Configuration**: Define KPIs, KRIs, categories, and scoring mechanisms using simple YAML files
- **Static Data Storage**: Store metrics in YAML files for simplicity and version control
- **Flexible Reporting**: Generate reports on overall posture or drill down into specific categories
- **Customizable Scoring**: Configure weights and thresholds for scoring at both global and category levels
- **Executive Levers**: Adjust scoring parameters to reflect organizational priorities

## Architecture

Pulse uses three main YAML configuration files:

1. **Metrics Configuration** (`config/metrics.yaml`): Defines the KPIs and KRIs organized by categories
2. **Metrics Data** (`data/metrics.yaml`): Contains the actual metric values (static data)
3. **Executive Levers** (`config/levers.yaml`): Defines scoring weights and thresholds

### Metrics Configuration Structure

```yaml
categories:
  - id: "app_sec"
    name: "Application Security"
    description: "Metrics related to application security posture"
    kpis:
      - id: "vuln_remediation_time"
        name: "Vulnerability Remediation Time"
        description: "Average time to remediate vulnerabilities"
        unit: "days"
        scoring_bands:
          - score: 95  # 95 points
            max: 15    # for values 0-15 days
          - score: 85  # 85 points
            min: 15    # for values 15-30 days
            max: 30
          - score: 75  # 75 points
            min: 30    # for values 30-45 days
            max: 45
          - score: 65  # 65 points
            min: 45    # for values 45-60 days
            max: 60
          - score: 30  # 30 points
            min: 60    # for values > 60 days
      # More KPIs...
    kris:
      - id: "critical_vulns"
        name: "Critical Vulnerabilities"
        description: "Number of critical vulnerabilities"
        unit: "count"
        scoring_bands:
          - score: 95  # 95 points
            max: 0     # for 0 vulnerabilities
          - score: 85  # 85 points
            min: 0     # for 1-2 vulnerabilities
            max: 2
          - score: 75  # 75 points
            min: 2     # for 3-5 vulnerabilities
            max: 5
          - score: 65  # 65 points
            min: 5     # for 6-10 vulnerabilities
            max: 10
          - score: 30  # 30 points
            min: 10    # for > 10 vulnerabilities
      # More KRIs...
  # More categories...
```

### Metrics Data Structure

```yaml
metrics:
  - reference: "app_sec.KPI.vuln_remediation_time"
    value: 45
    timestamp: "2025-04-01T00:00:00Z"
  - reference: "app_sec.KRI.critical_vulns"
    value: 3
    timestamp: "2025-04-01T00:00:00Z"
  # More metrics...
```

### Executive Levers Structure

```yaml
global:
  thresholds:
    green:
      min: 80  # Minimum score for green status
      max: 100 # Maximum score for green status
    yellow:
      min: 60  # Minimum score for yellow status
      max: 79  # Maximum score for yellow status
    red:
      min: 0   # Minimum score for red status
      max: 59  # Maximum score for red status
  
  # KPI and KRI specific thresholds
  kpi_thresholds:
    green:
      min: 85
      max: 100
    yellow:
      min: 65
      max: 84
    red:
      min: 0
      max: 64
  
  kri_thresholds:
    green:
      min: 75
      max: 100
    yellow:
      min: 55
      max: 74
    red:
      min: 0
      max: 54

weights:
  categories:
    "app_sec": 0.4  # Application Security
    "infra_sec": 0.3  # Infrastructure Security
    "compliance": 0.3  # Compliance
  
  # Category-specific thresholds (optional, overrides global)
  category_thresholds:
    "compliance":
      green:
        min: 85
        max: 100
      yellow:
        min: 70
        max: 84
      red:
        min: 0
        max: 69
```

## Installation

### Prerequisites

- Go 1.21 or higher

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/pulse.git
cd pulse

# Build the application
go build -o pulse cmd/main.go
```

## Usage

```bash
# Display overall posture
pulse report

# Display metrics for a specific category
pulse report --category "Application Security"

# Generate a detailed report in JSON format
pulse report --format json --output report.json

# Generate a report in tabular format
pulse report --format table

# Update metric values
pulse update --metric "app_sec.KRI.critical_vulns" --value 2

# List all available metrics
pulse list metrics

# List all categories
pulse list categories

# Display version information
pulse version
# or
pulse --version

# Initialize configuration files in the default location
pulse init

# Initialize configuration files in a specific directory
pulse init /path/to/directory
```

## Configuration

You can initialize the configuration files in two ways:

1. In the default location (~/.pulse/):
   ```bash
   pulse init
   ```

2. In a specific directory:
   ```bash
   pulse init /path/to/directory
   ```

This will create:
- `config/metrics.yaml`: Define your KPIs and KRIs organized by categories
- `config/levers.yaml`: Define scoring weights and thresholds
- `data/metrics/`: Directory containing metric files organized by category (e.g., `app_sec.yaml`)

You can then customize these files according to your needs.

## Development

### Project Structure

```
pulse/
├── cmd/
│   ├── main.go                      # CLI entry point
│   ├── cmd_root.go                  # Root command definition
│   ├── cmd_report.go                # Report command
│   ├── cmd_update.go                # Update command
│   ├── cmd_list.go                  # List command
│   ├── cmd_metrics.go               # Metrics command
│   ├── cmd_levers.go                # Levers command
│   ├── cmd_init.go                  # Init command
│   ├── cmd_version.go               # Version command
│   └── ...                          # Other command files
├── config.go                        # Configuration loading
├── metrics.go                       # Metrics processing
├── report.go                        # Report generation
├── score.go                         # Scoring calculations
├── types.go                         # Type definitions
├── constants.go                     # Constants
├── version.go                       # Version information
├── config/                          # Configuration directory
│   ├── metrics.yaml                 # Metrics configuration
│   └── levers.yaml                  # Executive levers configuration
├── data/                            # Data directory
│   └── metrics/                     # Metrics data directory
│       ├── app_sec.yaml             # Application security metrics
│       └── ...                      # Other metric files
├── go.mod
├── go.sum
└── README.md
```

The project is structured to be used both as a CLI application and as a library:
- The CLI artifact lives in the `cmd` directory
- The main business logic lives in the root directory (package: `pulse`)
- This allows other Go projects to import and use the `pulse` package as a library

### Adding New Features

1. Define new command in `cmd/`
2. Implement supporting logic in the root package
3. Update tests
4. Update documentation

### Building and Releasing

The project includes a Taskfile with various tasks for building, testing, and releasing:

```bash
# Build the application
task build

# Run tests
task test

# Generate test coverage report
task coverage

# Create release artifacts for Windows, Linux, and macOS
task release
```

The release task builds the application for multiple platforms (Windows, Linux, and macOS) and creates compressed archives in the `.dist` directory.

#### Version Information

The application includes version information that is set at build time:

- `version`: Determined from the current git branch or tag
- `build`: Determined from the git commit hash

When building a release, the version will be set to the git tag if the repository is checked out at a tag.

## License

[MIT](LICENSE)
