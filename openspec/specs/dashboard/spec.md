# Dashboard

## Purpose

Provide system overview and visualization of rule performance and alert trends.

## Requirements

### Requirement: System Overview Statistics

The system SHALL display key metrics on the dashboard:
- Total rule count
- Enabled rule count
- 24-hour alert count
- Elasticsearch data source status (success/total)

The statistics SHALL update automatically every 30 seconds.

#### Scenario: Display system stats
- **WHEN** user accesses dashboard
- **THEN** system displays all key metrics
- **AND** statistics are current (within 30 seconds)

#### Scenario: Auto refresh stats
- **WHEN** dashboard is open
- **THEN** statistics refresh every 30 seconds
- **AND** updates are reflected without page reload

### Requirement: Rule Alert Time Series Chart

The system SHALL display an interactive area chart showing alert trends over 24 hours.

The chart SHALL show data for all enabled rules.

The chart SHALL aggregate alerts by time buckets with intervals dynamically calculated based on each rule's execution frequency.

The chart SHALL display rules with no alerts as value 0.

The chart SHALL use different colors for each rule.

The chart SHALL be interactive with hover tooltips showing detailed values.

#### Scenario: Display time series chart
- **WHEN** user views dashboard
- **THEN** interactive area chart is displayed
- **AND** shows alert trends for last 24 hours
- **AND** includes all enabled rules

#### Scenario: Chart with no alerts
- **WHEN** a rule has no alerts in 24-hour period
- **THEN** chart shows value 0 for that rule across all time buckets
- **AND** rule is still visible in chart

#### Scenario: Interactive tooltip
- **WHEN** user hovers over chart data point
- **THEN** tooltip displays: rule name, time, and alert execution count
- **AND** tooltip is formatted clearly
- **AND** tooltip handles rules with different time bucket intervals correctly

### Requirement: Chart Data Aggregation

The system SHALL aggregate alert execution counts by time buckets.

The system SHALL use COUNT(*) for aggregation (counting alert records, not log entries).

The system SHALL calculate time bucket interval dynamically for each rule based on its execution interval.

The system SHALL use bucket interval of approximately 5x the rule's execution interval, rounded to common intervals (5, 15, 30, or 60 minutes).

The system SHALL generate time buckets for 24-hour period with rule-specific intervals.

The system SHALL fill missing buckets with value 0.

#### Scenario: Aggregate data with dynamic intervals
- **WHEN** alerts are generated at various times
- **THEN** system groups alerts into time buckets with interval based on each rule's execution frequency
- **AND** each bucket shows alert execution count (COUNT of alert records)
- **AND** buckets with no alerts show 0
- **AND** rules with shorter execution intervals use finer-grained buckets (e.g., 5 minutes)
- **AND** rules with longer execution intervals use coarser-grained buckets (e.g., 60 minutes)

#### Scenario: Fill missing buckets
- **WHEN** a rule has no alerts in certain time periods
- **THEN** chart displays 0 for those time buckets
- **AND** chart remains continuous

### Requirement: Chart Visualization

The chart SHALL use stacked area chart format.

The chart SHALL use distinct colors for each rule (up to 8 colors, then repeat).

The chart SHALL include:
- X-axis: Time labels (HH:MM format)
- Y-axis: Log count scale
- Legend: Rule names with colors
- Grid lines for readability

#### Scenario: Render chart
- **WHEN** chart data is loaded
- **THEN** chart renders with proper axes, labels, and legend
- **AND** colors are distinct and visually appealing

#### Scenario: Handle many rules
- **WHEN** dashboard has more than 8 enabled rules
- **THEN** chart uses color palette (repeating colors as needed)
- **AND** legend shows all rules

### Requirement: Data Loading and Caching

The dashboard SHALL load chart data from API endpoint.

The dashboard SHALL cache data for 30 seconds to reduce API calls.

The dashboard SHALL show loading indicator while fetching data.

The dashboard SHALL handle API errors gracefully.

#### Scenario: Load chart data
- **WHEN** dashboard loads
- **THEN** system fetches time-series data from API
- **AND** displays loading indicator
- **AND** renders chart when data arrives

#### Scenario: Cache data
- **WHEN** dashboard refreshes within 30 seconds
- **THEN** cached data is used
- **AND** no API call is made

#### Scenario: Handle API error
- **WHEN** chart data API call fails
- **THEN** system displays error message
- **AND** dashboard remains functional for other sections

### Requirement: Responsive Design

The dashboard SHALL be responsive and work on different screen sizes.

The chart SHALL adapt to container width.

The layout SHALL stack vertically on mobile devices.

#### Scenario: Desktop view
- **WHEN** dashboard is viewed on desktop
- **THEN** statistics cards and chart are displayed side-by-side or in grid
- **AND** chart uses full available width

#### Scenario: Mobile view
- **WHEN** dashboard is viewed on mobile
- **THEN** layout stacks vertically
- **AND** chart remains readable and interactive

