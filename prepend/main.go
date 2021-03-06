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
// Examples
//
// Prepends package.go with 4 first lines of license.go file:
//
//   ~ $ head -4 license.go | prepend package.go
//
// Prepends package.go with preamble.txt only if the file does not beging with
// it already:
//
//   ~ $ prepend -u -f preamble.txt package.go
package main

import (
	"bytes"
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

	prepend [-f INPUT_FILE] [-u] FILE...

EXAMPLE:

	Prepends package.go with 4 first lines of license.go file:

	  ~ $ head -4 license.go | prepend package.go

	Prepends package.go with preamble.txt only if the file does
	not beging with it already:

	  ~ $ prepend -u -f preamble.txt package.go`

var src string
var dst []string
var unique bool
var stdin bytes.Buffer

func nonil(err ...error) error {
	for _, err := range err {
		if err != nil {
			return err
		}
	}
	return nil
}

func die(v ...interface{}) {
	for _, v := range v {
		fmt.Fprintln(os.Stderr, v)
	}
	os.Exit(1)
}

func isfile(s string) error {
	if fi, err := os.Stat(s); err != nil && fi.IsDir() {
		return nonil(err, errors.New(s+" is a directory"))
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
	flag.BoolVar(&unique, "u", false, "")
	flag.BoolVar(&help, "help", false, "")
	flag.Parse()
	if help {
		flag.CommandLine.Usage()
	}
	switch flag.NArg() {
	case 0:
		die(usage)
	case 1:
		switch fi, err := os.Stat(flag.Arg(0)); {
		case err != nil:
			die(err)
		case fi.IsDir():
			f, err := os.Open(flag.Arg(0))
			if err != nil {
				die(err)
			}
			names, err := f.Readdirnames(-1)
			f.Close()
			for _, name := range names {
				dst = append(dst, filepath.Join(flag.Arg(0), name))
			}
		default:
			dst = append(dst, flag.Arg(0))
		}
	default:
		for _, s := range flag.Args() {
			if err := isfile(s); err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			dst = append(dst, s)
		}
	}

	// Early validate paths provided by the user.
	if src != "" {
		if err := isfile(src); err != nil {
			die(err)
		}
	}
}

var errNop = errors.New("nop")

type nopReader struct {
	r io.Reader
	n int
}

func (nr *nopReader) Read(p []byte) (int, error) {
	n, err := nr.r.Read(p)
	nr.n += n
	if err == io.EOF && nr.n == 0 {
		return 0, errNop
	}
	return n, err
}

func nop(r io.Reader) io.Reader {
	return &nopReader{r: r}
}

type uniqueReader struct {
	src    io.Reader
	dst    io.Reader
	bufsrc bytes.Buffer
	bufdst bytes.Buffer
	done   bool
	r      io.Reader
}

func (ur *uniqueReader) Read(p []byte) (int, error) {
	if ur.done {
		return ur.r.Read(p)
	}
	n, err := ur.src.Read(p)
	if n == 0 {
		return 0, errNop
	}
	q := make([]byte, n)
	m, e := ur.dst.Read(q)
	if m != n {
		ur.done = true
		return ur.r.Read(p)
	}
	for i := range q {
		if q[i] != p[i] {
			ur.done = true
			return ur.r.Read(p)
		}
	}
	switch {
	case err == io.EOF:
		return 0, errNop
	case e == io.EOF:
		ur.done = true
		return ur.r.Read(p[:n])
	default:
		return ur.r.Read(p[:n])
	}
}

func multiunique(src, dst io.Reader) io.Reader {
	ur := &uniqueReader{}
	ur.src = io.TeeReader(src, &ur.bufsrc)
	ur.dst = io.TeeReader(dst, &ur.bufdst)
	ur.r = io.MultiReader(&ur.bufsrc, src, &ur.bufdst, dst)
	return ur
}

func main() {
	var err []interface{}
	for _, dst := range dst {
		if e := prepend(dst); e != nil {
			err = append(err, e)
		}
	}
	if len(err) != 0 {
		die(err...)
	}
}

func prepend(dst string) (err error) {
	tmp, err := ioutil.TempFile(filepath.Split(dst))
	if err != nil {
		return err
	}
	rdst, err := os.Open(dst)
	if err != nil {
		return nonil(err, tmp.Close(), os.Remove(tmp.Name()))
	}
	var errCleanup error
	defer func() {
		switch errCleanup {
		case nil:
			if err = nonil(tmp.Close(), rdst.Close()); err != nil {
				os.Remove(tmp.Name())
			}
			// os.Rename fails under Windows if destination file exists.
			if err = os.Remove(dst); err != nil {
				os.Remove(tmp.Name())
			}
			if err = os.Rename(tmp.Name(), dst); err != nil {
				err = errors.New(err.Error() + " (prepended content is safe under " + tmp.Name() + ")")
			}
		default:
			nonil(tmp.Close(), rdst.Close(), os.Remove(tmp.Name()))
			if errCleanup != errNop {
				err = errCleanup
			}
		}
	}()
	var r io.Reader
	fi, err := os.Stdin.Stat()
	if err != nil {
		errCleanup = err
		return
	}
	switch {
	case src != "":
		f, err := os.Open(src)
		if err != nil {
			errCleanup = err
			return err
		}
		defer f.Close()
		r = f
	case fi.Mode()&os.ModeCharDevice == 0: // stackoverflow.com/questions/22744443
		r = io.MultiReader(bytes.NewReader(stdin.Bytes()), io.TeeReader(os.Stdin, &stdin))
	default:
		errCleanup = errNop
	}
	if unique {
		r = multiunique(nop(r), rdst)
	} else {
		r = io.MultiReader(nop(r), rdst)
	}
	_, errCleanup = io.Copy(tmp, r)
	return err
}
