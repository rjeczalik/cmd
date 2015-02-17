// Command dln prints derivative computed out of line-separated numbers.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

var errSyntax = errors.New("dln: line does not begin with a number")

func nonil(err ...error) error {
	for _, err := range err {
		if err != nil {
			return err
		}
	}
	return nil
}

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func main() {
	var r io.Reader = os.Stdin
	switch len(os.Args) {
	case 1:
	case 2:
		f, err := os.Open(os.Args[1])
		if err != nil {
			die(err)
		}
		defer f.Close()
		r = f
	default:
		die("dln: invalid arguments")
	}
	s := bufio.NewScanner(r)
	if !s.Scan() {
		return
	}
	var prev, cur int64
	if n, err := fmt.Sscanf(s.Text(), "%d", &prev); err != nil || n != 1 {
		die(nonil(err, errSyntax))
	}
	for s.Scan() {
		if n, err := fmt.Sscanf(s.Text(), "%d", &cur); err != nil || n != 1 {
			die(nonil(err, errSyntax))
		}
		fmt.Println(cur - prev)
		prev = cur
	}
	if err := s.Err(); err != nil {
		die(err)
	}
}
