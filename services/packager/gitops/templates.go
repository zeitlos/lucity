package gitops

import (
	"fmt"
	"time"
)

// projectYAML generates the project.yaml metadata file content.
func projectYAML(name, sourceURL string, createdAt time.Time) string {
	return fmt.Sprintf(`name: %s
source_url: %s
created_at: %s
`, name, sourceURL, createdAt.Format(time.RFC3339))
}

// baseChartYAML generates the base Chart.yaml that depends on lucity-app.
func baseChartYAML(project string) string {
	return fmt.Sprintf(`apiVersion: v2
name: %s
type: application
version: 0.1.0

dependencies:
  - name: lucity-app
    version: "0.1.0"
    repository: "file://../../charts/lucity-app"
`, project)
}

// baseValuesYAML generates the base values.yaml with empty service definitions.
const baseValuesYAML = `services: {}
`

// environmentValuesYAML generates the per-environment values.yaml override file.
const environmentValuesYAML = `# Environment-specific overrides
`
