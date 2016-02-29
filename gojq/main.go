package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var format = flag.String("t", "{{json .}}", "Format JSON with the given text/template.")
var text = flag.Bool("r", false, "Treat input text as text/plain.")
var input = flag.String("i", "-", "Input file; reads from stdin by default.")

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func isStdin() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice == 0
}

var funcs = template.FuncMap{
	"json": func(v interface{}) (string, error) {
		p, err := json.MarshalIndent(v, "", "\t")
		if err != nil {
			return "", err
		}
		return string(p), nil
	},
	"xml": func(v interface{}) (string, error) {
		p, err := xml.MarshalIndent(v, "", "\t")
		if err != nil {
			return "", err
		}
		return string(p), nil
	},
	"base": func(s string) string {
		return filepath.Base(s)
	},
	"dir": func(s string) string {
		return filepath.Dir(s)
	},
	"host": func(s string) (string, error) {
		u, err := url.Parse(s)
		if err == nil {
			s = u.Host
		}
		host, _, err := net.SplitHostPort(s)
		if err != nil {
			return s, nil
		}
		return host, nil
	},
}

func parse() (*template.Template, io.ReadCloser) {
	flag.Parse()
	if *format == "" {
		die("empty value passed for -f")
	}
	tmpl, err := template.New("gojq").Funcs(funcs).Parse(*format)
	if err != nil {
		die(err)
	}
	switch *input {
	case "":
		die("empty value passed for -i")
	case "-":
		if !isStdin() {
			die("no data on stdin")
		}
		return tmpl, ioutil.NopCloser(os.Stdin)
	}
	f, err := os.Open(*input)
	if err != nil {
		die(err)
	}
	return tmpl, f
}

func main() {
	var m interface{}
	tmpl, rc := parse()
	defer rc.Close()
	if *text {
		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if err := tmpl.Execute(os.Stdout, line); err != nil {
				die(err)
			}
			fmt.Println()
		}
		if err := scanner.Err(); err != nil {
			die(err)
		}
	} else {
		if err := json.NewDecoder(rc).Decode(&m); err != nil {
			die(err)
		}
		if err := tmpl.Execute(os.Stdout, m); err != nil {
			die(err)
		}
		fmt.Println()
	}
}
