// Command ph is a pipe helper.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

var (
	batchNum  = flag.Int("b", 1, "Aggregates input lines in batches before passing on.")
	batchJoin = flag.String("j", ",", "Character to join lines for a batch.")
	sleep     = flag.Duration("sleep", 0, "Time to sleep between batches.")
)

var output = func(batch []string) {
	var buf bytes.Buffer
	buf.WriteString(batch[0])
	for _, s := range batch[1:] {
		buf.WriteString(*batchJoin)
		buf.WriteString(s)
	}
	buf.WriteByte('\n')
	io.Copy(os.Stdout, &buf)
}

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func main() {
	flag.Parse()
	if *batchNum < 1 {
		*batchNum = 1
	}
	var cmd []string
	for i, arg := range flag.Args() {
		cmd = flag.Args()[:i+1]
		if arg == "--" {
			break
		}
	}
	if len(cmd) != 0 {
		output = func(batch []string) {
			var buf bytes.Buffer
			for _, s := range batch {
				fmt.Fprintln(&buf, s)
			}
			c := exec.Command(cmd[0], cmd[1:]...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = &buf
			if err := c.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
	if err := run(); err != nil {
		die(err)
	}
}

func run() error {
	batch := make([]string, *batchNum)
	current := 0
	scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
	for scanner.Scan() {
		batch[current] = scanner.Text()
		current++
		if current == len(batch) {
			output(batch)
			current = 0
			if *sleep != 0 {
				time.Sleep(*sleep)
			}
		}
	}
	output(batch[:current])
	return scanner.Err()
}
