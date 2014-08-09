// Command gotree is a reimplmentation of the Unix tree command in Go.
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

	"github.com/rjeczalik/tools/fs"
	"github.com/rjeczalik/tools/fs/fsutil"
	"github.com/rjeczalik/tools/fs/memfs"
)

const usage = `NAME:
	gotree - Go implementation of the Unix tree command

USAGE:
	gotree [OPTION]... [DIRECTORY]

OPTIONS:
	-a			Lists also hidden files
	-d			Lists directories only (NOT IMPLEMENTED)
	-L level	Descends only <level> directories deep`

var (
	all bool
	dir bool
	lvl int
)

func init() {
	flag.BoolVar(&all, "a", false, "")
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
	var root string
	if len(flag.Args()) > 1 {
		die(usage)
	}
	if len(flag.Args()) == 1 {
		root = filepath.Clean(flag.Args()[0])
	} else {
		root, _ = os.Getwd()
	}
	var spy = memfs.New()
	(fsutil.Control{FS: fsutil.TeeFilesystem(fs.FS{}, spy), Hidden: all}).Find(root, lvl)
	spy, err := spy.Cd(root)
	if err != nil {
		// TODO(rjeczalik): improve error message
		die(err)
	}
	var ndir, nfile int
	spy.Walk(string(os.PathSeparator), func(_ string, fi os.FileInfo, _ error) (err error) {
		if fi.IsDir() {
			ndir++
		} else {
			nfile++
		}
		return
	})
	// Root directory does not count.
	ndir--
	fmt.Printf("%s%c%s\n%d directories, %d files\n", root, os.PathSeparator, spy, ndir, nfile)
}
