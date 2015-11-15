package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
)

var (
	baseURL   = flag.String("url", "https://gcr.io/v1/", "Docker Repository endpoint.")
	basicUser = flag.String("user", "", "Username (read from ~/.dockercfg when empty)")
	basicPass = flag.String("pass", "", "Password (read from ~/.dockercfg when empty)")
	image     = flag.String("image", "", "Docker image name - gcr.io/{image} (required)")
)

type Tag struct {
	Tag   string `json:"tag"`
	Image string `json:"image"`
}

func NewTags(m registry.TagMap) []Tag {
	tags := make([]Tag, 0, len(m))
	for tag, image := range m {
		tags = append(tags, Tag{Tag: tag, Image: image})
	}
	sort.Sort(byTagDesc(tags))
	return tags
}

type Config map[string]struct {
	Auth []byte `json:"auth"`
}

func (cfg Config) UserPass(host string) (user, pass string, ok bool) {
	for k := range cfg {
		u, err := url.Parse(k)
		if err != nil {
			continue
		}
		if u.Host != host {
			continue
		}
		auth := string(cfg[k].Auth)
		if i := strings.IndexRune(auth, ':'); i != -1 {
			return auth[:i], auth[i+1:], true
		}
	}
	return "", "", false
}

func ReadConfig(file string) (Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cfg := make(Config)
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

func userpass(host string) (usr, pass string) {
	u, err := user.Current()
	if err != nil {
		die("unable to read current user:", err)
	}
	file := filepath.Join(u.HomeDir, ".dockercfg")
	cfg, err := ReadConfig(file)
	if err != nil {
		die("unable to read .dockercfg:", err)
	}
	usr, pass, ok := cfg.UserPass(host)
	if !ok {
		die("no auth found for", host, "(try gcloud docker --authorize-only)")
	}
	return usr, pass
}

func main() {
	flag.Parse()
	if *image == "" {
		die("image name is empty or missing")
	}
	u, err := url.Parse(*baseURL)
	if err != nil {
		die("unable to parse baseURL:", err)
	}
	if *basicUser == "" || *basicPass == "" {
		*basicUser, *basicPass = userpass(u.Host)
	}
	c := registry.NewClient()
	c.BaseURL = u
	basic := registry.BasicAuth{
		Username: *basicUser,
		Password: *basicPass,
	}
	token, err := c.Hub.GetReadTokenWithAuth(*image, basic)
	if err != nil {
		die("failed to obtain read token:", err)
	}
	tags, err := c.Repository.ListTags(*image, token)
	if err != nil {
		die("failed to obtain tag list:", err)
	}
	p, err := json.MarshalIndent(NewTags(tags), "", "\t")
	if err != nil {
		die("failed to JSON encode tags:", err)
	}
	fmt.Printf("%s\n", p)
}

type byTagDesc []Tag

func (p byTagDesc) Len() int           { return len(p) }
func (p byTagDesc) Less(i, j int) bool { return p[i].Tag > p[j].Tag }
func (p byTagDesc) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
