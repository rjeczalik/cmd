// Command gotree is a reimplmentation of the Unix tree command in Go.
//
// It is not going to be a feature-complete drop-in replacement for the tree
// command, but rather a showcase of fs/memfs and fs/fsutil packages.
//
// Usage
//
//   NAME:
//     gotree - Go implementation of the Unix tree command
//
//   USAGE:
//     gotree [OPTION]... [DIRECTORY]
//
//   OPTIONS:
//     -a          All files are listed
//     -d          List directories only
//     -L  level   Descend only <level> directories deep
//     -go width   Output as Go literal with specified maximum column width
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
	"io"
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
	-a             Lists also hidden files
	-d             Lists directories only
	-L   level     Descends only <level> directories deep
	-go  width     Output tree as Go literal with the specified column width
	-var name      Output tree as Go variable with the specified name
	               (if not otherwise specified, column width is set to 80)`

var (
	all     bool
	dir     bool
	lvl     int
	gowidth int
	varname string
)

var flags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func init() {
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
	}
	flags.BoolVar(&all, "a", false, "")
	flags.BoolVar(&dir, "d", false, "")
	flags.IntVar(&lvl, "L", 0, "")
	flags.IntVar(&gowidth, "go", 0, "")
	flags.StringVar(&varname, "var", "", "")
	flags.Parse(os.Args[1:])
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
	if len(flags.Args()) == 1 {
		root = flags.Args()[0]
	}
	if root == "." {
		root, _ = os.Getwd()
		printroot = false
	}
	root = filepath.Clean(root)
	(fsutil.Control{FS: fsutil.TeeFilesystem(fs.FS{}, spy), Hidden: all}).Find(root, lvl)
	spy, err := spy.Cd(root)
	if err != nil {
		die(err) // TODO(rjeczalik): improve error message
	}
	if gowidth > 0 || varname != "" {
		if err = EncodeLiteral(spy, gowidth, varname, os.Stdout); err != nil {
			die(err)
		}
	} else {
		if err = gotree(root, printroot, spy, os.Stdout); err != nil {
			die(err)
		}
	}
}

func gotree(root string, printroot bool, spy memfs.FS, w io.Writer) (err error) {
	var (
		r      io.Reader
		pr, pw = io.Pipe()
		ch     = make(chan error, 1)
		ndir   int
		nfile  int
		fn     filepath.WalkFunc
	)
	if dir {
		fn = countdirdelfile(&ndir, spy)
	} else {
		fn = countdirfile(&ndir, &nfile)
	}
	if err = spy.Walk(string(os.PathSeparator), fn); err != nil {
		return
	}
	go func() {
		ch <- nonnil(memfs.Unix.Encode(spy, pw), pw.Close())
	}()
	switch {
	case dir && printroot:
		r = io.MultiReader(
			strings.NewReader(fmt.Sprintf("%s%c", root, os.PathSeparator)),
			pr,
			strings.NewReader(fmt.Sprintf("\n%d directories\n", ndir-1)),
		)
	case dir:
		r = io.MultiReader(
			pr,
			strings.NewReader(fmt.Sprintf("\n%d directories\n", ndir-1)),
		)
	case printroot:
		r = io.MultiReader(
			strings.NewReader(fmt.Sprintf("%s%c", root, os.PathSeparator)),
			pr,
			strings.NewReader(fmt.Sprintf("\n%d directories, %d files\n", ndir-1, nfile)),
		)
	default:
		r = io.MultiReader(
			pr,
			strings.NewReader(fmt.Sprintf("\n%d directories, %d files\n", ndir-1, nfile)),
		)
	}
	_, err = io.Copy(w, r)
	if e := <-ch; e != nil && err == nil {
		err = e
	}
	return
}
