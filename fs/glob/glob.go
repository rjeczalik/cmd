// Package glob is a collection of functions useful for traversing a filesystem
// in all unusual ways.
package glob

import (
	"path/filepath"

	"github.com/rjeczalik/tools/fs"
)

// Readdirnames reads all names of all subdirectories of the 'dir', except
// the ones which begin with a dot.
func Readdirnames(dir string) []string {
	return defaultGlob.Readdirnames(dir)
}

// Intersect returns a collection of paths which are the longest intersection
// between two directory tries - those tries have roots in 'src' and 'dir' directories.
// It does not glob into directories, which names begin with a dot.
//
// Example
//
// For the following filesystem:
//
//   .
//   ├── data
//   │   └── github.com
//   │       └── user
//   │           └── example
//   │               └── assets
//   │                   ├── css
//   │                   └── js
//   └── src
//       └── github.com
//           └── user
//               └── example
//
// The following call:
//
//   names := glob.Intersect("src", "data")
//
// Gives:
//
//   []string{"github.com/user/example"}
func Intersect(src, dir string) []string {
	return defaultGlob.Intersect(src, dir)
}

// Glob is the glob package's control structure, allows for changing the behavior
// of its functions.
type Glob struct {
	// FS specifies the mechanism using which Glob accesses the filesystem.
	FS fs.Filesystem
	// Hidden tells whether the files and directories which name begin with a dot
	// should be included in the results.
	Hidden bool
}

// Readdirnames reads all names of all subdirectories of the 'dir'.
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

// Intersect returns a collection of paths which are the longest intersection
// between two directory tries - those tries have roots in 'src' and 'dir' directories.
func (g Glob) Intersect(src, dir string) []string {
	glob, dirs, pop := []string{""}, map[string]struct{}{"": {}}, ""
	for len(glob) > 0 {
		pop, glob = glob[len(glob)-1], glob[:len(glob)-1]
		subdir := g.Readdirnames(filepath.Join(dir, pop))
		if subdir == nil {
			dirs[pop] = struct{}{}
			continue
		}
		subsrc := g.Readdirnames(filepath.Join(src, pop))
		if subsrc == nil {
			dirs[pop] = struct{}{}
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
			dirs[pop] = struct{}{}
		}
	}
	delete(dirs, "")
	if len(dirs) == 0 {
		return nil
	}
	s := make([]string, 0, len(dirs))
	for k := range dirs {
		s = append(s, k)
	}
	return s
}

func (g Glob) hidden(name string) bool {
	return !g.Hidden && name[0] == '.'
}

var defaultGlob = Glob{
	FS:     fs.Default,
	Hidden: false,
}
