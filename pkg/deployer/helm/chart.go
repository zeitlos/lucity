package helm

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"helm.sh/helm/v4/pkg/chart/loader/archive"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
)

func LoadChartFromFS(fsys fs.FS, root string) (*chart.Chart, error) {
	var files []*archive.BufferedFile

	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(fsys, p)

		if err != nil {
			return err
		}

		rel := strings.TrimPrefix(p, root+"/")
		rel = path.Clean(rel)

		files = append(files, &archive.BufferedFile{
			Name: rel,
			Data: data,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk chart fs at %s: %w", root, err)
	}

	return loader.LoadFiles(files)
}
