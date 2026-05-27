package softserve

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/zeitlos/lucity/charts"
	"github.com/zeitlos/lucity/services/conductor/internal/deployerold/argo/gitops"
	"gopkg.in/yaml.v3"
)

// subchartKey is the Helm dependency name used in GitOps repos.
// Values must be scoped under this key for Helm to pass them to the subchart.
const subchartKey = "lucity-app"

// baseChartYAML generates the base Chart.yaml that depends on lucity-app.
func baseChartYAML(project string) string {
	return fmt.Sprintf(`apiVersion: v2
name: %s
type: application
version: 0.1.0

dependencies:
  - name: lucity-app
    version: "0.1.0"
    repository: "file://../chart"
`, project)
}

// baseValuesYAML generates the base values.yaml with empty service definitions.
// Values are scoped under the "lucity-app" key because the chart is a subchart
// dependency — Helm requires subchart values to be namespaced this way.
// fullnameOverride is set to the project name so K8s resource names are concise
// (e.g., "beast-web" instead of "acme-beast-dev-lucity-app-web").
//
// workspace and project are emitted as top-level values so the chart can
// stamp them onto every rendered resource via lucity-app.labels — the
// platform package's discovery selectors depend on these labels.
func baseValuesYAML(workspace, project string) string {
	return fmt.Sprintf(`lucity-app:
  fullnameOverride: "%s"
  workspace: "%s"
  project: "%s"
  services: {}
`, project, workspace, project)
}

// environmentValuesYAML generates the per-environment values.yaml override file.
// environment is emitted so the chart can label every resource with
// lucity.dev/environment for platform discovery.
func environmentValuesYAML(environment string) string {
	return fmt.Sprintf(`lucity-app:
  environment: "%s"
`, environment)
}

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

func parseServiceDefs(services map[string]any) []gitops.ServiceDef {
	var result []gitops.ServiceDef
	for svcName, svcRaw := range services {
		svcMap, ok := svcRaw.(map[string]any)
		if !ok {
			continue
		}

		def := gitops.ServiceDef{Name: svcName}

		if imageMap, ok := svcMap["image"].(map[string]any); ok {
			if repo, ok := imageMap["repository"].(string); ok {
				def.Image = repo
			}
		}
		if port, ok := svcMap["port"].(int); ok {
			def.Port = port
		}
		if framework, ok := svcMap["framework"].(string); ok {
			def.Framework = framework
		}
		if sourceURL, ok := svcMap["sourceUrl"].(string); ok {
			def.SourceURL = sourceURL
		}
		if contextPath, ok := svcMap["contextPath"].(string); ok {
			def.ContextPath = contextPath
		}
		if installStr, ok := svcMap["githubInstallationId"].(string); ok {
			if id, err := strconv.ParseInt(installStr, 10, 64); err == nil {
				def.GitHubInstallationID = id
			}
		} else if installID, ok := svcMap["githubInstallationId"].(int); ok {
			// Legacy: handle values written as int before the string fix.
			def.GitHubInstallationID = int64(installID)
		}
		if cmd, ok := svcMap["customStartCommand"].(string); ok {
			def.CustomStartCommand = cmd
		}
		if cmd, ok := svcMap["startCommand"].(string); ok {
			def.StartCommand = cmd
		}

		result = append(result, def)
	}
	return result
}

// parseServiceInstanceMetas extracts service metadata from a YAML services map.
func parseServiceInstanceMetas(services map[string]any) []gitops.ServiceInstanceMeta {
	var result []gitops.ServiceInstanceMeta
	for svcName, svcRaw := range services {
		svcMap, ok := svcRaw.(map[string]any)
		if !ok {
			continue
		}

		meta := gitops.ServiceInstanceMeta{Name: svcName}
		if imageMap, ok := svcMap["image"].(map[string]any); ok {
			if tag, ok := imageMap["tag"].(string); ok {
				meta.ImageTag = tag
			}
		}
		if domains, ok := svcMap["domains"].([]any); ok {
			for _, d := range domains {
				if s, ok := d.(string); ok {
					meta.Domains = append(meta.Domains, s)
				}
			}
		}

		result = append(result, meta)
	}
	return result
}

// parseDatabaseDefs extracts database definitions from a YAML values map.
func parseDatabaseDefs(inner map[string]any) []gitops.DatabaseDef {
	databases, ok := inner["databases"].(map[string]any)
	if !ok {
		return nil
	}
	postgres, ok := databases["postgres"].(map[string]any)
	if !ok {
		return nil
	}

	var result []gitops.DatabaseDef
	for dbName, dbRaw := range postgres {
		dbMap, ok := dbRaw.(map[string]any)
		if !ok {
			continue
		}
		def := gitops.DatabaseDef{Name: dbName}
		if v, ok := dbMap["version"].(string); ok {
			def.Version = v
		}
		if v, ok := dbMap["instances"].(int); ok {
			def.Instances = v
		}
		if v, ok := dbMap["size"].(string); ok {
			def.Size = v
		}
		result = append(result, def)
	}
	return result
}

// parseDatabaseRefs extracts database references from a YAML service map.
func parseDatabaseRefs(m map[string]any) map[string]gitops.DatabaseRef {
	raw, ok := m["databaseRefs"].(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]gitops.DatabaseRef, len(raw))
	for k, v := range raw {
		ref, ok := v.(map[string]any)
		if !ok {
			continue
		}
		result[k] = gitops.DatabaseRef{
			Database: fmt.Sprintf("%v", ref["database"]),
			Key:      fmt.Sprintf("%v", ref["key"]),
		}
	}
	return result
}

// databaseRefsToAny converts database refs to map[string]any for YAML marshaling.
func databaseRefsToAny(refs map[string]gitops.DatabaseRef) map[string]any {
	result := make(map[string]any, len(refs))
	for k, v := range refs {
		result[k] = map[string]any{
			"database": v.Database,
			"key":      v.Key,
		}
	}
	return result
}

// parseServiceRefs extracts service references from a YAML service map.
func parseServiceRefs(m map[string]any) map[string]gitops.ServiceRef {
	raw, ok := m["serviceRefs"].(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]gitops.ServiceRef, len(raw))
	for k, v := range raw {
		ref, ok := v.(map[string]any)
		if !ok {
			continue
		}
		result[k] = gitops.ServiceRef{
			Service: fmt.Sprintf("%v", ref["service"]),
		}
	}
	return result
}

// serviceRefsToAny converts service refs to map[string]any for YAML marshaling.
func serviceRefsToAny(refs map[string]gitops.ServiceRef) map[string]any {
	result := make(map[string]any, len(refs))
	for k, v := range refs {
		result[k] = map[string]any{
			"service": v.Service,
		}
	}
	return result
}

// readLocalValuesYAML reads and parses a local YAML file.
func readLocalValuesYAML(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var values map[string]any
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	if values == nil {
		values = make(map[string]any)
	}

	return values, nil
}

// readSubchartValues reads the lucity-app subchart values from a local values.yaml.
func readSubchartValues(path string) (map[string]any, error) {
	values, err := readLocalValuesYAML(path)

	if err != nil {
		return nil, err
	}

	inner, ok := values[subchartKey].(map[string]any)
	if !ok {
		inner = make(map[string]any)
	}

	return inner, nil
}

// readSubchartValuesFromBytes parses raw YAML bytes and extracts the subchart values.
// TODO: This should probably make use of readSubchartValues
func readSubchartValuesFromBytes(data []byte) (map[string]any, error) {
	var values map[string]any
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("failed to parse values: %w", err)
	}
	if values == nil {
		values = make(map[string]any)
	}
	inner, ok := values[subchartKey].(map[string]any)
	if !ok {
		inner = make(map[string]any)
	}
	return inner, nil
}

// writeLocalValuesYAML marshals values and writes them to a local file.
func writeLocalValuesYAML(path string, values map[string]any) error {
	data, err := yaml.Marshal(values)

	if err != nil {
		return fmt.Errorf("failed to marshal values: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// writeSubchartValues writes values nested under the subchart key.
func writeSubchartValues(path string, inner map[string]any) error {
	return writeLocalValuesYAML(path, map[string]any{subchartKey: inner})
}
