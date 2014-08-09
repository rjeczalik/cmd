// Command mktree creates a file tree out of tree output read from standard input.
// The support format is output from gotree or Unix tree commands.
//
// Usage
//
//   NAME:
//   	mktree - mirrors a file tree when pipped to gotree command
//
//   USAGE:
//   	mktree [-h] [-o=.|dir] [-d] [<file>]
//
//   OPTIONS:
//   	-o <dir>  Destination directory - defaults to current directory
//   	-d        Create directories only
//   	-h        Print usage information
//   	<file>    Read tree output from the file instead of standard input
//
// Example
//
// The following example mirrors file tree of tools/fs package under /tmp/mktree
// directory.
//
//   src/github.com/rjeczalik/tools/fs $ gotree
//   .
//   ├── fs.go
//   ├── fsutil
//   │   ├── fsutil.go
//   │   ├── fsutil_test.go
//   │   ├── tee.go
//   │   └── tee_test.go
//   └── memfs
//       ├── memfs.go
//       ├── memfs_test.go
//       ├── tree.go
//       ├── tree_test.go
//       ├── util.go
//       └── util_test.go
//
//   2 directories, 11 files
//   src/github.com/rjeczalik/tools/fs $ gotree | mktree -o /tmp/mktree
//   src/github.com/rjeczalik/tools/fs $ gotree /tmp/mktree
//   /tmp/mktree/.
//   ├── fs.go
//   ├── fsutil
//   │   ├── fsutil.go
//   │   ├── fsutil_test.go
//   │   ├── tee.go
//   │   └── tee_test.go
//   └── memfs
//       ├── memfs.go
//       ├── memfs_test.go
//       ├── tree.go
//       ├── tree_test.go
//       ├── util.go
//       └── util_test.go
//
//   2 directories, 11 files
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rjeczalik/tools/fs"
	"github.com/rjeczalik/tools/fs/memfs"
)

const usage = `NAME:
	mktree - mirrors a file tree when pipped to gotree command

USAGE:
	mktree [-h] [-o=.|dir] [-d] [<file>]

OPTIONS:
	-o <dir>  Destination directory - defaults to current directory
	-d        Create directories only
	-h        Print usage information
	<file>    Read tree output from the file instead of standard input`

var (
	dironly bool
	output  string
)

var flags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func init() {
	output, _ = os.Getwd()
	flags.StringVar(&output, "o", output, "")
	flags.BoolVar(&dironly, "d", false, "")
	flags.Usage = func() { fmt.Println(usage) }
	flags.Parse(os.Args[1:])
}

func main() {
	if len(flags.Args()) > 1 {
		die(usage)
	}
	var in io.Reader = os.Stdin
	if len(flags.Args()) == 1 {
		f, err := os.Open(flags.Args()[0])
		if err != nil {
			die(err)
		}
		defer f.Close()
		in = f
	}
	tree, err := memfs.Unix.Decode(in)
	if err != nil {
		die(err)
	}
	fn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dst := filepath.Join(output, path)
		dstfi, err := os.Stat(dst)
		if err == nil {
			if fi.IsDir() != dstfi.IsDir() {
				err = fmt.Errorf("create: %s already exists", dst)
			}
			return err
		}
		if base := filepath.Join(output, filepath.Dir(path)); base != dst {
			// TODO(rjeczalik): dir mode, not base
			if err = fs.MkdirAll(base, fi.Mode()); err != nil {
				return err
			}
		}
		if fi.IsDir() {
			err = fs.Mkdir(dst, fi.Mode())
		} else {
			var f fs.File
			if f, err = fs.Create(dst); err == nil {
				err = f.Close()
			}
		}
		return err
	}
	if err = tree.Walk(string(os.PathSeparator), fn); err != nil {
		die(err)
	}
}
