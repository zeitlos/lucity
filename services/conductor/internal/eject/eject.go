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

const (
	releaseName = "lucity-app"
	docsURL     = "https://lucity.cloud/docs/eject#installing-on-your-own-cluster"
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
	fmt.Fprintf(&b, "Before the first install, read <%s>. It covers what your cluster needs, which values to change, and how to move your data.\n\n", docsURL)
	fmt.Fprintf(&b, "Each environment is a separate Helm release in its own namespace. The release must be named `%s`, because the values refer to resources by names derived from it. Install one with:\n\n", releaseName)
	b.WriteString("```sh\n")

	if len(envs) > 0 {
		fmt.Fprintf(&b, "helm upgrade --install %s ./chart \\\n", releaseName)
		fmt.Fprintf(&b, "  -f values/%s.yaml \\\n", envs[0].Name)
		fmt.Fprintf(&b, "  --namespace %s-%s --create-namespace\n", project.Name, envs[0].Name)
	} else {
		fmt.Fprintf(&b, "helm upgrade --install %s ./chart \\\n", releaseName)
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
	b.WriteString("- **Container images**: services built from source point at Lucity's internal registry, which your cluster cannot reach. Rebuild them, push them to a registry your cluster can pull from, and update `image` in the values, dropping the old `digest`.\n")
	b.WriteString("- **Image pull secret**: if your images are private, create the pull secret referenced under `imagePullSecrets` in your target namespace.\n")
	b.WriteString("- **Gateway**: HTTP routing expects a Gateway API gateway. Point the `gateway` values at one you run, or remove the routes if you front traffic differently.\n")
	b.WriteString("- **Databases**: PostgreSQL clusters use the CloudNativePG operator. Install it before deploying, or adjust the database values to match your setup.\n")
	b.WriteString("- **Buckets**: services that use a bucket read its credentials from a secret named `lucity-bucket-<bucket>`. Create it in your target namespace.\n\n")

	b.WriteString("Databases, key-value stores and volumes start out empty. Your project keeps running on Lucity. This export is a copy, not a migration.\n")

	return []byte(b.String())
}
