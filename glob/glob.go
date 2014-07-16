// Package glob
package glob

import (
	"path/filepath"

	"github.com/rjeczalik/tools/fs"
)

// Readdirnames
func Readdirnames(dir string) []string {
	return Default.Readdirnames(dir)
}

// Intersect
func Intersect(src, dir string) []string {
	return Default.Intersect(src, dir)
}

// Glob
type Glob struct {
	FS     fs.Filesystem //
	Hidden bool          //
}

// Readdirnames
func (g Glob) Readdirnames(dir string) []string {
	f, err := g.FS.Open(dir)
	if err != nil {
		return nil
	}
	defer f.Close()
	fi, err := f.Readdir(0)
	if err != nil || len(fi) == 0 {
		return nil
	}
	names := make([]string, 0, len(fi))
	for _, fi := range fi {
		if fi.IsDir() {
			name := filepath.Base(fi.Name())
			if !g.hidden(name) {
				names = append(names, name)
			}
		}
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

// Gopath
func (g Glob) Gopath() []string {
	return nil
}

// Intersect
func (g Glob) Intersect(src, dir string) []string {
	glob, dirs, pop := []string{""}, []string{}, ""
	for len(glob) > 0 {
		pop, glob = glob[len(glob)-1], glob[:len(glob)-1]
		subdir := g.Readdirnames(filepath.Join(dir, pop))
		if subdir == nil {
			if pop != "" {
				dirs = append(dirs, pop)
			}
			continue
		}
		subsrc := g.Readdirnames(filepath.Join(src, pop))
		if subsrc == nil {
			if pop != "" {
				dirs = append(dirs, pop)
			}
			continue
		}
	LOOP:
		for i := range subdir {
			for j := range subsrc {
				if subdir[i] == subsrc[j] {
					glob = append(glob, filepath.Join(pop, subdir[i]))
					continue LOOP
				}
			}
			if pop != "" {
				dirs = append(dirs, pop)
			}
		}
	}
	if len(dirs) == 0 {
		return nil
	}
	return dirs
}

func (g Glob) hidden(name string) bool {
	return !g.Hidden && name[0] == '.'
}

// Default
var Default = Glob{
	FS:     fs.Default,
	Hidden: false,
}
