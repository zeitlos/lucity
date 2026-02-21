package eject

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/zeitlos/lucity/charts"
	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/packager/gitops"

	"gopkg.in/yaml.v3"
)

// Build produces a zip archive of the ejected project.
// It reads all files from the GitOps repo, bundles the embedded Helm chart,
// generates ArgoCD manifests, a build script, and a README.
func Build(ctx context.Context, provider gitops.Provider, project string) ([]byte, error) {
	shortName := labels.ShortName(project)
	prefix := shortName + "-ejected/"

	// Read raw files from the GitOps repo.
	repoFiles, err := provider.RepoFiles(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo files: %w", err)
	}

	// Extract environment names and service info from the repo files.
	environments := environmentsFromFiles(repoFiles)
	services := servicesFromFiles(repoFiles)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// 1. Write GitOps repo files (project.yaml, base/*, environments/*).
	for path, content := range repoFiles {
		if err := writeZipFile(zw, prefix+path, content); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	// 2. Write the embedded lucity-app chart under chart/.
	if err := writeEmbeddedChart(zw, prefix); err != nil {
		return nil, fmt.Errorf("failed to write embedded chart: %w", err)
	}

	// 3. Generate ArgoCD Application manifests.
	for _, env := range environments {
		content := applicationYAML(project, env)
		appName := labels.NamespaceFor(project, env)
		path := fmt.Sprintf("argocd/applications/%s.yaml", appName)
		if err := writeZipFile(zw, prefix+path, []byte(content)); err != nil {
			return nil, fmt.Errorf("failed to write ArgoCD manifest for %s: %w", env, err)
		}
	}

	// 4. Generate build script.
	script := buildScript(services)
	if err := writeZipFile(zw, prefix+"build.sh", []byte(script)); err != nil {
		return nil, fmt.Errorf("failed to write build.sh: %w", err)
	}

	// 5. Generate README.
	readme := readmeContent(project, environments, services)
	if err := writeZipFile(zw, prefix+"README.md", []byte(readme)); err != nil {
		return nil, fmt.Errorf("failed to write README.md: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip: %w", err)
	}

	return buf.Bytes(), nil
}

// writeZipFile creates a file entry in the zip archive.
func writeZipFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// writeEmbeddedChart walks the embedded lucity-app chart FS and writes
// each file to chart/ in the zip archive.
func writeEmbeddedChart(zw *zip.Writer, prefix string) error {
	return fs.WalkDir(charts.LucityApp, "lucity-app", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Map "lucity-app/..." to "chart/..."
		rel := strings.TrimPrefix(path, "lucity-app/")
		data, err := fs.ReadFile(charts.LucityApp, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		return writeZipFile(zw, prefix+"chart/"+rel, data)
	})
}

// environmentsFromFiles extracts environment names from the repo file map.
// Looks for paths matching environments/<name>/values.yaml.
func environmentsFromFiles(files map[string][]byte) []string {
	var envs []string
	seen := make(map[string]bool)
	for path := range files {
		if !strings.HasPrefix(path, "environments/") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(path, "environments/"), "/", 2)
		if len(parts) >= 1 && parts[0] != "" && !seen[parts[0]] {
			envs = append(envs, parts[0])
			seen[parts[0]] = true
		}
	}
	return envs
}

// servicesFromFiles parses base/values.yaml to extract service definitions
// for build script generation.
func servicesFromFiles(files map[string][]byte) []serviceInfo {
	data, ok := files["base/values.yaml"]
	if !ok {
		return nil
	}

	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil
	}

	// Values are scoped under lucity-app: key (subchart convention).
	inner, ok := root["lucity-app"].(map[string]any)
	if !ok {
		return nil
	}

	servicesMap, ok := inner["services"].(map[string]any)
	if !ok {
		return nil
	}

	var services []serviceInfo
	for name, v := range servicesMap {
		svc, ok := v.(map[string]any)
		if !ok {
			continue
		}
		si := serviceInfo{Name: name}

		if imgMap, ok := svc["image"].(map[string]any); ok {
			if repo, ok := imgMap["repository"].(string); ok {
				si.Image = repo
			}
		}

		services = append(services, si)
	}
	return services
}
