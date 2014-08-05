// Package fsutil is a collection of various filesystem utility functions.
package fsutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rjeczalik/tools/fs"
)

// Readpaths reads names of all the files and directories of the 'dir' directory.
// If none files were found, the 'files' slice will be nil. If none directories
// were found, the 'dirs' slice will be nil. If the 'dir' was empty or error
// occured, both slice will be empty.
func Readpaths(dir string) (files, dirs []string) {
	return defaultControl.Readpaths(dir)
}

// Readdirpaths reads all names of all subdirectories of the 'dir', except
// the ones which begin with a dot.
func Readdirpaths(dir string) []string {
	return defaultControl.Readdirpaths(dir)
}

// Intersect returns a collection of paths which are the longest intersection
// between two directory trees - those trees have roots in 'src' and 'dir' directories.
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
	return defaultControl.Intersect(src, dir)
}

// Find globs into 'dir' directory, reading all files and directories except those,
// which names begin with a dot.
//
// For n > 0 it descends for n directories deep.
// For n <= 0 it reads all directories.
//
// On success it returns full paths for files and directories it found.
func Find(dir string, n int) []string {
	return defaultControl.Find(dir, n)
}

// Control is the package control structure, allows for altering the behavior
// of its functions.
type Control struct {
	// FS specifies the mechanism using which Glob accesses the filesystem.
	FS fs.Filesystem
	// Hidden tells whether the files and directories which name begin with a dot
	// should be included in the results.
	Hidden bool
}

// Readpaths reads names of all the files and directories of the 'dir' directory.
// If none files were found, the 'files' slice will be nil. If none directories
// were found, the 'dirs' slice will be nil. If the 'dir' was empty or error
// occured, both slice will be empty.
func (c Control) Readpaths(dir string) (files, dirs []string) {
	return c.readall(dir)
}

// Readdirpaths reads names of all the subdirectories of the 'dir' directory.
// If none were found or error occured, returned slice is nil.
func (c Control) Readdirpaths(dir string) []string {
	_, d := c.readall(dir)
	return d
}

func (c Control) readall(dir string) (files, dirs []string) {
	f, err := c.FS.Open(dir)
	if err != nil {
		return nil, nil
	}
	defer f.Close()
	fi, err := f.Readdir(0)
	if err != nil || len(fi) == 0 {
		return nil, nil
	}
	for _, fi := range fi {
		if name := filepath.Base(fi.Name()); !c.hidden(name) {
			if fi.IsDir() {
				dirs = append(dirs, name)
			} else {
				files = append(files, name)
			}
		}
	}
	if len(files) == 0 {
		files = nil
	}
	if len(dirs) == 0 {
		dirs = nil
	}
	return
}

func depthBelow(depth int, root, dir string) bool {
	if depth <= 0 {
		return true
	}
	return strings.Count(dir[strings.Index(dir, root)+len(root):],
		string(os.PathSeparator)) < depth
}

// Find globs into 'dir' directory, reading all files and directories.
//
// For n > 0 it descends for n directories deep.
// For n <= 0 it reads all directories.
//
// On success it returns full paths for files and directories it found.
func (c Control) Find(dir string, n int) []string {
	var (
		path string
		all  []string
		glob = []string{dir}
	)
	for len(glob) > 0 {
		path, glob = glob[len(glob)-1], glob[:len(glob)-1]
		files, dirs := c.Readpaths(path)
		for _, file := range files {
			all = append(all, filepath.Join(path, filepath.Base(file)))
		}
		for _, d := range dirs {
			d = filepath.Join(path, filepath.Base(d))
			if depthBelow(n, dir, d) {
				glob = append(glob, d)
			}
			all = append(all, d)
		}
	}
	if len(all) == 0 {
		return nil
	}
	return all
}

// Intersect returns a collection of paths which are the longest intersection
// between two directory trees - those trees have roots in 'src' and 'dir' directories.
func (c Control) Intersect(src, dir string) []string {
	glob, dirs, pop := []string{""}, map[string]struct{}{"": {}}, ""
	for len(glob) > 0 {
		pop, glob = glob[len(glob)-1], glob[:len(glob)-1]
		subdir := c.Readdirpaths(filepath.Join(dir, pop))
		if subdir == nil {
			dirs[pop] = struct{}{}
			continue
		}
		subsrc := c.Readdirpaths(filepath.Join(src, pop))
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

func (c Control) hidden(name string) bool {
	return !c.Hidden && name[0] == '.'
}

var defaultControl = Control{
	FS:     fs.Default,
	Hidden: false,
}
