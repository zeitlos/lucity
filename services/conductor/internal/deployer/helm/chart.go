package helm

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func LoadChartFromFS(fsys fs.FS, root string) (*chart.Chart, error) {
	var files []*loader.BufferedFile

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

		files = append(files, &loader.BufferedFile{
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
