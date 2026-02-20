package gitops

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeitlos/lucity/charts"
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
	// Helm chart names can't contain slashes — use only the short name.
	name := project
	if parts := strings.SplitN(project, "/", 2); len(parts) == 2 {
		name = parts[1]
	}
	return fmt.Sprintf(`apiVersion: v2
name: %s
type: application
version: 0.1.0

dependencies:
  - name: lucity-app
    version: "0.1.0"
    repository: "file://../chart"
`, name)
}

// baseValuesYAML generates the base values.yaml with empty service definitions.
// Values are scoped under the "lucity-app" key because the chart is a subchart
// dependency — Helm requires subchart values to be namespaced this way.
const baseValuesYAML = `lucity-app:
  services: {}
`

// environmentValuesYAML generates the per-environment values.yaml override file.
const environmentValuesYAML = `lucity-app: {}
`

// writeEmbeddedChart writes the embedded lucity-app chart to a "chart/" directory
// inside the given root directory. Used during GitOps repo initialization so that
// ArgoCD can resolve the chart dependency locally.
func writeEmbeddedChart(rootDir string) error {
	return fs.WalkDir(charts.LucityApp, "lucity-app", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Map "lucity-app/..." to "chart/..."
		rel, err := filepath.Rel("lucity-app", path)
		if err != nil {
			return err
		}
		target := filepath.Join(rootDir, "chart", rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := fs.ReadFile(charts.LucityApp, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		return os.WriteFile(target, data, 0o644)
	})
}
