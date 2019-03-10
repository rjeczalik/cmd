# gotmpl

Command line tool for executing Go's templates. It templates a file with input data (JSON- or YAML-encoded) writing result to stdout.

### Usage

```
USAGE:

	gotmpl TEMPLATE_FILE|- [DATA_FILE.json|DATA_FILE.yaml|-]


```

### Example


```bash
$ cat >input.yaml <<EOF
> data:
>   value: "foo"
>   enabled: true
> EOF
```
```bash
$ cat >template.tmpl <<EOF
> {{- if .data.enabled }}
> The {{ .data.value }} feature is enabled.
> {{- else }}
> The {{ .data.value }} feature is disabled.
> {{- end }}
> EOF
```
```bash
 $ gotmpl template.tmpl input.yaml 
The foo feature is enabled.
```
