package sourcetree

import (
	"io/fs"
	"os"
	"path/filepath"
)

type symlink struct {
	path   string
	target string
}

// HideEscapingSymlinks removes the symlinks under dir that do not resolve to a
// file inside dir, dangling ones included, so tools that read the tree cannot
// be pointed at files elsewhere on the host. The returned function puts them
// back.
func HideEscapingSymlinks(dir string) (func() error, error) {
	root, err := os.OpenRoot(dir)

	if err != nil {
		return nil, err
	}

	defer root.Close()

	var hidden []symlink

	err = fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.Type()&fs.ModeSymlink == 0 {
			return nil
		}

		if _, err := root.Stat(path); err == nil {
			return nil
		}

		target, err := root.Readlink(path)

		if err != nil {
			return err
		}

		hidden = append(hidden, symlink{path: filepath.Join(dir, path), target: target})

		return root.Remove(path)
	})

	if err != nil {
		return nil, err
	}

	restore := func() error {
		for _, link := range hidden {
			if err := os.Symlink(link.target, link.path); err != nil {
				return err
			}
		}

		return nil
	}

	return restore, nil
}
