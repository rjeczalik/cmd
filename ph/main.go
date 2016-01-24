// Command ph is a pipe helper.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	batchNum = flag.Int("b", 1, "Aggregates input lines in batches before passing on.")
	sleep    = flag.Duration("sleep", 0, "Time to sleep between batches.")
)

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func main() {
	flag.Parse()
	if *batchNum < 1 {
		*batchNum = 1
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
			for _, s := range batch {
				fmt.Println(s)
			}
			current = 0
			if *sleep != 0 {
				time.Sleep(*sleep)
			}
		}
	}
	for _, s := range batch[:current] {
		fmt.Println(s)
	}
	return scanner.Err()
}
