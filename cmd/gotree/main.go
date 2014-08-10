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
//   	-a          All files are listed
//   	-d          List directories only
//   	-L  level   Descend only <level> directories deep
//      -go width   Output as Go literal with specified maximum column width
//
// Example
//
//   ~/src $ gotree -a -L 1 github.com/rjeczalik/tools
//   github.com/rjeczalik/tools/.
//   ├── .git/
//   ├── .gitignore
//   ├── .travis.yml
//   ├── LICENSE
//   ├── README.md
//   ├── appveyor.yml
//   ├── cmd/
//   ├── doc.go
//   ├── fs/
//   ├── netz/
//   └── rw/
//
//   5 directories, 6 files
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
	-a            Lists also hidden files
	-d            Lists directories only
	-L  level     Descends only <level> directories deep
	-go width     Output as Go literal with specified maximum column width`

var (
	all     bool
	dir     bool
	lvl     int
	gowidth int
)

func init() {
	flag.BoolVar(&all, "a", false, "")
	flag.BoolVar(&dir, "d", false, "")
	flag.IntVar(&lvl, "L", 0, "")
	flag.IntVar(&gowidth, "go", 0, "")
	flag.Parse()
}

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func ishelp(s string) bool {
	return s == "-h" || s == "-help" || s == "help" || s == "--help" || s == "/?"
}

func countdirfile(ndir, nfile *int) filepath.WalkFunc {
	return func(_ string, fi os.FileInfo, _ error) (err error) {
		if fi.IsDir() {
			*ndir++
		} else {
			*nfile++
		}
		return
	}
}

func countdirdelfile(ndir *int, fs memfs.FS) filepath.WalkFunc {
	return func(s string, fi os.FileInfo, _ error) (err error) {
		if fi.IsDir() {
			*ndir++
		} else {
			err = fs.Remove(s)
		}
		return
	}
}

func main() {
	if len(os.Args) == 2 && ishelp(os.Args[1]) {
		fmt.Println(usage)
		return
	}
	if len(flag.Args()) > 1 {
		die(usage)
	}
	var (
		root      = "."
		spy       = memfs.New()
		printroot = true
	)
	if len(flag.Args()) == 1 {
		root = flag.Args()[0]
	}
	if root == "." {
		root, _ = os.Getwd()
		printroot = false
	}
	root = filepath.Clean(root)
	(fsutil.Control{FS: fsutil.TeeFilesystem(fs.FS{}, spy), Hidden: all}).Find(root, lvl)
	spy, err := spy.Cd(root)
	if err != nil {
		// TODO(rjeczalik): improve error message
		die(err)
	}
	var (
		ndir  int
		nfile int
		fn    filepath.WalkFunc
	)
	if dir {
		fn = countdirdelfile(&ndir, spy)
	} else {
		fn = countdirfile(&ndir, &nfile)
	}
	if err = spy.Walk(string(os.PathSeparator), fn); err != nil {
		die(err)
	}
	switch {
	case gowidth > 0:
		if err = EncodeLiteral(spy, gowidth, os.Stdout); err != nil {
			die(err)
		}
	case dir && printroot:
		fmt.Printf("%s%c%s\n%d directories\n", root, os.PathSeparator, spy, ndir-1)
	case dir:
		fmt.Printf("%s\n%d directories\n", spy, ndir-1)
	case printroot:
		fmt.Printf("%s%c%s\n%d directories, %d files\n", root, os.PathSeparator, spy, ndir-1, nfile)
	default:
		fmt.Printf("%s\n%d directories, %d files\n", spy, ndir-1, nfile)
	}
}
