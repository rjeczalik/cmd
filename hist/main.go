// Command hist prints histogram for line-separated data points.
//
// TODO(rjeczalik): make sorting optional and disabled by default
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

func min(i, j int) int {
	if i > j {
		return j
	}
	return i
}

type pair struct {
	s string
	n int
}

type byvalue []pair // implements counting set ordered by value

func (b byvalue) Len() int           { return len(b) }
func (b byvalue) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byvalue) Less(i, j int) bool { return b[i].s < b[j].s }

func (b byvalue) Search(s string) int {
	return sort.Search(len(b), func(i int) bool { return b[i].s >= s })
}

func (b *byvalue) Add(s string) {
	switch i := b.Search(s); {
	case i == len(*b):
		*b = append(*b, pair{s: s, n: 1})
	case (*b)[i].s == s:
		(*b)[i].n++
	default:
		*b = append(*b, pair{})
		copy((*b)[i+1:], (*b)[i:])
		(*b)[i] = pair{s: s, n: 1}
	}
}

type bycount []pair

func (b bycount) Len() int           { return len(b) }
func (b bycount) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b bycount) Less(i, j int) bool { return b[i].n >= b[j].n }
func (b bycount) Sort()              { sort.Sort(b) }

var errSyntax = errors.New("invalid range value syntax")

type sliceVar [2]int

func (s *sliceVar) Set(str string) error {
	m := strings.Split(str, ":")
	if len(m) != 2 {
		return errSyntax
	}
	switch {
	case m[0] == "" && m[1] == "":
		*s = [2]int{0, -1}
	case m[0] == "":
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return err
		}
		*s = [2]int{0, n}
	case m[1] == "" || m[1] == "-1":
		n, err := strconv.Atoi(m[0])
		if err != nil {
			return err
		}
		*s = [2]int{n, -1}
	default:
		n, err := strconv.Atoi(m[0])
		if err != nil {
			return err
		}
		o, err := strconv.Atoi(m[1])
		if err != nil {
			return err
		}
		*s = [2]int{n, o}
	}
	return nil
}

func (s sliceVar) String() string {
	switch {
	case s == [2]int{0, -1}:
		return ":"
	case s[0] == 0:
		return ":" + strconv.Itoa(s[1])
	case s[1] == -1:
		return strconv.Itoa(s[0]) + ":"
	default:
		return strconv.Itoa(s[0]) + ":" + strconv.Itoa(s[1])
	}
}

var slice = sliceVar{0, -1}

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func init() {
	flag.Var(&slice, "slice", "limit result set using given slice indices")
	flag.Parse()
}

func main() {
	var r io.Reader = os.Stdin
	switch flag.NArg() {
	case 0:
	case 1:
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			die(err)
		}
		defer f.Close()
		r = f
	default:
		die("hist: invalid arguments")
	}
	max := 0
	hist := byvalue{}
	s := bufio.NewScanner(r)
	for s.Scan() {
		s := s.Text()
		if n := len(s); n > max {
			max = n
		}
		hist.Add(s)
	}
	if err := s.Err(); err != nil {
		die(err)
	}
	bycount(hist).Sort()
	// Slice the result set.
	//
	// TODO(rjeczalik): update max
	n := len(hist)
	if slice[1] == -1 {
		slice[1] = n
	}
	hist = hist[min(slice[0], n):min(slice[1], n)]
	if len(hist) == 0 {
		return
	}
	format := "%" + strconv.Itoa(max+1) + "s\t%d\t"
	n, _ = fmt.Printf(format, hist[0].s, hist[0].n)
	n = 70 - n
	fmt.Println(strings.Repeat("░", n))
	for i := 1; i < len(hist); i++ {
		fmt.Printf(format, hist[i].s, hist[i].n)
		if r := n * hist[i].n / hist[0].n; r != 0 {
			fmt.Print(strings.Repeat("░", r))
		}
		fmt.Println()
	}
}
