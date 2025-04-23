# Pulse

A Risk and Performance measurement framework CLI application for security programs.

## Overview

Pulse is a command-line tool designed to report on Key Performance Indicators (KPIs) and Key Risk Indicators (KRIs) for security programs. It provides a flexible framework for defining, measuring, and reporting on security metrics across multiple categories.

## Features

- **YAML-based Configuration**: Define KPIs, KRIs, categories, and scoring mechanisms using simple YAML files
- **Static Data Storage**: Store metrics in YAML files for simplicity and version control
- **Flexible Reporting**: Generate reports on overall security posture or drill down into specific categories
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
        target: 30
        scoring_bands:
          band_5: 15  # 0-15 days: 90-100 points
          band_4: 30  # 16-30 days: 80-89 points
          band_3: 45  # 31-45 days: 70-79 points
          band_2: 60  # 46-60 days: 60-69 points
          band_1: 61  # 61+ days: 0-59 points
      # More KPIs...
    kris:
      - id: "critical_vulns"
        name: "Critical Vulnerabilities"
        description: "Number of critical vulnerabilities"
        unit: "count"
        threshold: 5
        scoring_bands:
          band_5: 0   # 0 vulns: 90-100 points
          band_4: 2   # 1-2 vulns: 80-89 points
          band_3: 5   # 3-5 vulns: 70-79 points
          band_2: 10  # 6-10 vulns: 60-69 points
          band_1: 11  # 11+ vulns: 0-59 points
      # More KRIs...
  # More categories...
```

### Metrics Data Structure

```yaml
metrics:
  - reference: "app_sec.KPI.vuln_remediation_time"
    value: 45
    timestamp: "2025-04-01"
  - reference: "app_sec.KRI.critical_vulns"
    value: 3
    timestamp: "2025-04-01"
  # More metrics...
```

### Executive Levers Structure

```yaml
global:
  thresholds:
    green: 80  # 80-100 points
    yellow: 60 # 60-79 points
    red: 0     # 0-59 points
  
  # Scoring bands - 5 ranges within each threshold on a 100-point scale
  scoring_bands:
    band_5: 90  # 90-100 points
    band_4: 80  # 80-89 points
    band_3: 70  # 70-79 points
    band_2: 60  # 60-69 points
    band_1: 0   # 0-59 points
  
weights:
  categories:
    "app_sec": 0.4  # Application Security
    "infra_sec": 0.3  # Infrastructure Security
    "compliance": 0.3  # Compliance
  
  # Category-specific thresholds (optional, overrides global)
  category_thresholds:
    "compliance":
      green: 85
      yellow: 70
      red: 0
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
go build -o pulse cmd/pulse/main.go
```

## Usage

```bash
# Display overall security posture
pulse report

# Display metrics for a specific category
pulse report --category "Application Security"

# Generate a detailed report in JSON format
pulse report --format json --output report.json

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
- `data/metrics.yaml`: Store your metric values

You can then customize these files according to your needs.

## Development

### Project Structure

```
pulse/
├── cmd/
│   └── pulse/
│       └── main.go       # CLI entry point
├── config.go             # Configuration loading
├── metrics.go            # Metrics processing
├── report.go             # Report generation
├── score.go              # Scoring calculations
├── types.go              # Type definitions
├── config/
│   ├── metrics.yaml
│   └── levers.yaml
├── data/
│   └── metrics.yaml
├── go.mod
├── go.sum
└── README.md
```

The project is structured to be used both as a CLI application and as a library:
- The CLI artifact lives in the `cmd` directory
- The main business logic lives in the root directory (package: `pulse`)
- This allows other Go projects to import and use the `pulse` package as a library

### Adding New Features

1. Define new command in `cmd/pulse/`
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

[Include your license information here]
