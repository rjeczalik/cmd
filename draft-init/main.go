package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type ServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

type RegistryAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type options struct {
	file   string
	domain string
	gcr    string
	docker bool
	dry    bool
}

func (opts *options) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&opts.file, "f", "", "Service account JSON key file.")
	f.StringVar(&opts.domain, "domain", "", "Base domain for applications.")
	f.StringVar(&opts.gcr, "gcr", "gcr.io", "Container registry host.")
	f.BoolVar(&opts.dry, "dry", false, "Print commands only.")
	f.BoolVar(&opts.docker, "docker", false, "Initialize docker only.")
}
func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

func main() {
	var opts options
	opts.RegisterFlags(flag.CommandLine)
	flag.Parse()

	if err := run(&opts); err != nil {
		die(err)
	}
}

func run(opts *options) error {
	pass, err := ioutil.ReadFile(opts.file)
	if err != nil {
		return err
	}

	var cert ServiceAccount
	if err := json.Unmarshal(pass, &cert); err != nil {
		return err
	}
	if pass, err = json.Marshal(&cert); err != nil {
		return err
	}

	auth := &RegistryAuth{
		Username: "_json_key",
		Password: string(pass),
		Email:    "not@val.id",
	}

	p, err := json.Marshal(auth)
	if err != nil {
		return err
	}

	token := base64.StdEncoding.EncodeToString(p)

	draftConfig := []string{
		"registry.url=" + opts.gcr,
		"registry.org=" + cert.ProjectID,
		"registry.authtoken=" + token,
	}

	if opts.domain != "" {
		draftConfig = append(draftConfig, "basedomain="+opts.domain)
	}

	draftArgs := []string{
		"init",
		"--set", strings.Join(draftConfig, ","),
	}

	dockerArgs := []string{
		"login",
		"-e", auth.Email,
		"-u", auth.Username,
		"-p", auth.Password,
		opts.gcr,
	}

	draft := cmd("draft", draftArgs...)
	docker := cmd("docker", dockerArgs...)

	if opts.dry {
		fmt.Fprintln(os.Stderr, docker.Args)
		fmt.Fprintln(os.Stderr, draft.Args)
		return nil
	}

	if err := docker.Run(); err != nil {
		return err
	}

	if opts.docker {
		return nil
	}

	return draft.Run()
}

func cmd(name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c
}
