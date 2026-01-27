## MODIFIED Requirements

### Requirement: Chart Visualization

The chart SHALL use stacked area chart format.

The chart SHALL use distinct colors for each rule (up to 8 colors, then repeat).

The chart SHALL include:
- X-axis: Time labels showing the actual rolling 24-hour window from current time (e.g., "14:30" yesterday to "14:30" today), not day-aligned times
- Y-axis: Alert execution count scale
- Legend: Rule names with colors
- Grid lines for readability

**The X-axis time labels SHALL reflect the actual data range: from current time minus 24 hours to current time, not aligned to day boundaries (00:00-23:00).**

#### Scenario: Render chart
- **WHEN** chart data is loaded
- **THEN** chart renders with proper axes, labels, and legend
- **AND** colors are distinct and visually appealing
- **AND** X-axis shows rolling 24-hour window (e.g., "14:30" to "14:30" next day)

#### Scenario: Chart time range display
- **WHEN** user views chart at 14:30:45
- **THEN** X-axis displays time labels from approximately 14:30 yesterday to 14:30 today
- **AND** first data point label shows time near "14:30" (24 hours ago)
- **AND** last data point label shows time near "14:30" (current time)
- **AND** labels do NOT show day-aligned times like "00:00" to "23:00"

#### Scenario: Handle many rules
- **WHEN** dashboard has more than 8 enabled rules
- **THEN** chart uses color palette (repeating colors as needed)
- **AND** legend shows all rules
- **AND** X-axis still shows rolling 24-hour window correctly
