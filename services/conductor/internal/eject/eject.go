package eject

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

type Project struct {
	Name string
	ID   string
}

type EnvValues struct {
	Name   string
	Values []byte
}

func Build(chartFS fs.FS, project Project, envs []EnvValues) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	root := project.Name + "-ejected"

	err := fs.WalkDir(chartFS, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(chartFS, p)

		if err != nil {
			return err
		}

		return writeFile(zw, path.Join(root, "chart", p), data)
	})

	if err != nil {
		return nil, fmt.Errorf("walk chart: %w", err)
	}

	sort.Slice(envs, func(i, j int) bool { return envs[i].Name < envs[j].Name })

	for _, env := range envs {
		if err := writeFile(zw, path.Join(root, "values", env.Name+".yaml"), env.Values); err != nil {
			return nil, err
		}
	}

	if err := writeFile(zw, path.Join(root, "README.md"), readme(project, envs)); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close archive: %w", err)
	}

	return buf.Bytes(), nil
}

func writeFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)

	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}

	return nil
}

func readme(project Project, envs []EnvValues) []byte {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", project.Name)
	b.WriteString("This is a self-contained export of your Lucity project. It is a standard Helm chart plus one values file per environment, with no dependency on the Lucity control plane.\n\n")

	b.WriteString("## Layout\n\n")
	b.WriteString("```\n")
	b.WriteString("chart/      the lucity-app Helm chart\n")
	b.WriteString("values/     one values file per environment\n")
	b.WriteString("```\n\n")

	b.WriteString("## Deploy an environment\n\n")
	b.WriteString("Each environment is a separate Helm release. Install one with:\n\n")
	b.WriteString("```sh\n")

	if len(envs) > 0 {
		fmt.Fprintf(&b, "helm upgrade --install %s ./chart \\\n", project.Name)
		fmt.Fprintf(&b, "  -f values/%s.yaml \\\n", envs[0].Name)
		fmt.Fprintf(&b, "  --namespace %s-%s --create-namespace\n", project.Name, envs[0].Name)
	} else {
		fmt.Fprintf(&b, "helm upgrade --install %s ./chart \\\n", project.Name)
		fmt.Fprintf(&b, "  -f values/<environment>.yaml \\\n")
		fmt.Fprintf(&b, "  --namespace <namespace> --create-namespace\n")
	}

	b.WriteString("```\n\n")

	if len(envs) > 0 {
		b.WriteString("Environments in this export:\n\n")
		for _, env := range envs {
			fmt.Fprintf(&b, "- `%s`\n", env.Name)
		}
		b.WriteString("\n")
	}

	b.WriteString("## What you need to provide\n\n")
	b.WriteString("The values reflect exactly what ran on Lucity, so they reference infrastructure the platform provided for you. On your own cluster you supply the equivalents:\n\n")
	b.WriteString("- **Container images**: the `image` references point at the registry that built your workloads. Make sure your cluster can pull them, or rebuild and repoint the references.\n")
	b.WriteString("- **Image pull secret**: if your images are private, create the pull secret referenced under `imagePullSecrets` in your target namespace.\n")
	b.WriteString("- **Gateway**: HTTP routing expects a Gateway API gateway. Point the `gateway` values at one you run, or remove the routes if you front traffic differently.\n")
	b.WriteString("- **Databases**: PostgreSQL clusters use the CloudNativePG operator. Install it before deploying, or adjust the database values to match your setup.\n\n")

	b.WriteString("Your project keeps running on Lucity. This export is a copy, not a migration.\n")

	return []byte(b.String())
}
