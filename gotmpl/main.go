package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
)

const usage = `USAGE:

	gotmpl TEMPLATE_FILE [DATA_FILE.json]

`

func die(v ...interface{}) {
	log.Println(v...)
	os.Exit(1)
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stderr)

	if flag.NArg() != 2 {
		die(usage)
	}

	p, err := run(flag.Arg(0), flag.Arg(1))
	if err != nil {
		die(err)
	}

	os.Stdout.Write(p)
}

var f = map[string]interface{}{
	"json": func(v interface{}) (string, error) {
		p, err := json.MarshalIndent(v, "", "\t")
		if err != nil {
			return "", err
		}
		return string(p), nil
	},
	"or": func(s ...string) string {
		for _, s := range s {
			if s != "" {
				return s
			}
		}
		return ""
	},
	"file": func(path string) (string, error) {
		if path == "" {
			return "", nil
		}
		p, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(bytes.TrimSpace(p)), nil
	},
	"base64": func(s string) string {
		return base64.RawStdEncoding.EncodeToString([]byte(s))
	},
	"jsonl": func(v ...interface{}) (string, error) {
		var buf bytes.Buffer
		for _, v := range v {
			p, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			buf.Write(p)
			buf.WriteRune('\n')
		}
		return strings.TrimSpace(buf.String()), nil
	},
	"fromjson": func(s string) (interface{}, error) {
		var v interface{}
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			return nil, err
		}
		return v, nil
	},
	"quote": func(s string) string {
		return strconv.Quote(s)
	},
}

func run(tmplFile, dataFile string) ([]byte, error) {
	if tmplFile == "-" && dataFile == "-" {
		return nil, errors.New("template file and data file cannot be both stdin")
	}
	p, err := readFile(tmplFile)
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("gotmpl").Funcs(f).Parse(string(p))
	if err != nil {
		return nil, err
	}
	var data interface{}
	p, err = readFile(dataFile)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(p, &data); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readFile(path string) ([]byte, error) {
	if path == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(path)
}
