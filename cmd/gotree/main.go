// cmd/gotree is a reimplmentation of the Unix tree command in Go.
//
// It is not going to be a feature-complete drop-in replacement for the tree
// command, but rather a showcase of fs/memfs and fs/fsutil packages.
//
// Usage
//
//   NAME:
//   	gotree - Go implementation of the Unix tree command
//
//   USAGE:
//   	gotree [OPTION]... [DIRECTORY]
//
//   OPTIONS:
//   	-a          All files are listed (NOT IMPLEMENTED)
//   	-d          List directories only (NOT IMPLEMENTED)
//   	-L level    Descend only <level> directories deep
//
// Example
//
//   ~/src $ gotree -L 1 github.com/rjeczalik/tools
//   github.com
//   └── rjeczalik
//       └── tools
//           ├── .git/
//           ├── .gitignore
//           ├── .travis.yml
//           ├── LICENSE
//           ├── README.md
//           ├── cmd/
//           ├── doc.go
//           ├── fs/
//           └── netz/
//
//   7 directories, 5 files
//
// TODO
//
// * do not list hidden files (currently gotree has -a set by default)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rjeczalik/tools/fs"
	"github.com/rjeczalik/tools/fs/fsutil"
	"github.com/rjeczalik/tools/fs/memfs"
)

const usage = `NAME:
	gotree - Go implementation of the Unix tree command

USAGE:
	gotree [OPTION]... [DIRECTORY]

OPTIONS:
	-a			Lists also hidden files (NOT IMPLEMENTED)
	-d			Lists directories only (NOT IMPLEMENTED)
	-L level	Descends only <level> directories deep`

var (
	all bool
	dir bool
	lvl int
)

func init() {
	flag.BoolVar(&all, "a", false, "") // TODO
	flag.BoolVar(&dir, "d", false, "") // TODO
	flag.IntVar(&lvl, "L", 0, "")
	flag.Parse()
}

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func ishelp(s string) bool {
	return s == "-h" || s == "-help" || s == "help" || s == "--help" || s == "/?"
}

func main() {
	if len(os.Args) == 2 && ishelp(os.Args[1]) {
		fmt.Println(usage)
		return
	}
	var path = "."
	if len(flag.Args()) > 1 {
		die(usage)
	}
	if len(flag.Args()) == 1 {
		path = flag.Args()[0]
	}
	var (
		glob = []string{path}
		spy  = memfs.New()
		fs   = fsutil.TeeFilesystem(fs.FS{}, spy)
	)
	islvlok := func(path string) func(s string) bool {
		if lvl == 0 {
			return func(string) bool { return true }
		}
		return func(s string) bool {
			return strings.Count(s[strings.Index(s, path)+len(path):],
				string(os.PathSeparator)) < lvl
		}
	}(path)
	for len(glob) > 0 {
		path, glob = glob[len(glob)-1], glob[:len(glob)-1]
		f, err := fs.Open(path)
		if err != nil {
			continue
		}
		fi, err := f.Readdir(0)
		if err != nil {
			f.Close()
			continue
		}
		for _, fi := range fi {
			s := filepath.Join(path, filepath.Base(fi.Name()))
			if fi.IsDir() && islvlok(s) {
				glob = append(glob, s)
			}
		}
		f.Close()
	}
	var ndir, nfile int
	spy.Walk(".", func(_ string, fi os.FileInfo, _ error) (err error) {
		if fi.IsDir() {
			ndir++
		} else {
			nfile++
		}
		return
	})
	fmt.Printf("%s\n%d directories, %d files\n", spy, ndir, nfile)
}
