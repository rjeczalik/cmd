// Command prepend inserts data read from stdin or an input file at the begining
// of the given file.
//
// If data to prepend is passed both via stdin and input file, first the given
// file is prepended with data read from stdin, then from input file.
//
// The prepend command does not load the files contents to the memory, making
// it suitable for large files. Writes issued by the prepend command are atomic,
// meaning if reading from stdin or input file fails the original file is left
// untouched.
//
// Example
//
//  ~ $ head -4 license.go | prepend package.go
//
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const usage = `prepend - inserts data at the begining of the file

USAGE:

	prepend [-f INPUT_FILE] FILE

EXAMPLE:

	~ $ head -4 license.go | prepend package.go`

var src string
var dst string

func nonil(err ...error) error {
	for _, err := range err {
		if err != nil {
			return err
		}
	}
	return nil
}

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

func isfile(s string) error {
	if fi, err := os.Stat(dst); err != nil && fi.IsDir() {
		return nonil(err, errors.New(dst+" is a directory"))
	}
	return nil
}

func init() {
	flag.CommandLine.Usage = func() {
		fmt.Println(usage)
		os.Exit(0)
	}
	help := false
	flag.StringVar(&src, "f", "", "")
	flag.BoolVar(&help, "help", false, "")
	flag.Parse()
	if help {
		flag.CommandLine.Usage()
	}
	if flag.NArg() != 1 {
		die(usage)
	}
	dst = flag.Arg(0)
	// Early validate paths provided by the user.
	if err := isfile(dst); err != nil {
		die(err)
	}
	if src != "" {
		if err := isfile(src); err != nil {
			die(err)
		}
	}
}

func main() {
	tmp, err := ioutil.TempFile(filepath.Split(dst))
	if err != nil {
		die(err)
	}
	rdst, err := os.Open(dst)
	if err != nil {
		die(nonil(err, tmp.Close(), os.Remove(tmp.Name())))
	}
	var errCleanup error
	defer func() {
		switch errCleanup {
		case nil:
			if err = nonil(tmp.Close(), rdst.Close()); err != nil {
				os.Remove(tmp.Name())
				die(err)
			}
			// os.Renamve fails under Windows if destination file exists.
			if err = os.Remove(dst); err != nil {
				os.Remove(tmp.Name())
				die(err)
			}
			if err = os.Rename(tmp.Name(), dst); err != nil {
				die(err, "Prepended content is safe under ", tmp.Name())
			}
		default:
			die(nonil(errCleanup, tmp.Close(), rdst.Close(), os.Remove(tmp.Name())))
		}
	}()
	var r io.Reader = rdst
	if src != "" {
		f, err := os.Open(src)
		if err != nil {
			errCleanup = err
			return
		}
		defer f.Close()
		r = io.MultiReader(f, r)
	}
	// stackoverflow.com/questions/22744443
	fi, err := os.Stdin.Stat()
	if err != nil {
		errCleanup = err
		return
	}
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		r = io.MultiReader(os.Stdin, r)
	}
	_, errCleanup = io.Copy(tmp, r)
}
