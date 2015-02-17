tools [![Build Status](https://img.shields.io/travis/rjeczalik/tools/master.svg)](https://travis-ci.org/rjeczalik/tools "linux_amd64") [![Build Status](https://img.shields.io/travis/rjeczalik/tools/osx.svg)](https://travis-ci.org/rjeczalik/tools "darwin_amd64") [![Build status](https://img.shields.io/appveyor/ci/rjeczalik/tools-161.svg)](https://ci.appveyor.com/project/rjeczalik/tools-161 "windows_amd64") [![Coverage Status](https://img.shields.io/coveralls/rjeczalik/tools/master.svg)](https://coveralls.io/r/rjeczalik/tools?branch=master)
=====

Handmade tools for day-to-day plumbing.

## cmd/notify [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/cmd/notify?status.png)](https://godoc.org/github.com/rjeczalik/tools/cmd/notify)

Listens on filesystem changes and forwards received mapping to user-defined handlers.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/tools/cmd/notify
```

*Documentation*

[godoc.org/github.com/rjeczalik/tools/cmd/notify](http://godoc.org/github.com/rjeczalik/tools/cmd/notify)

*Usage*

```
~ $ notify -c 'echo "Hello from handler! (event={{.Event}}, path={{.Path}})"'
2015/02/17 01:17:40 received notify.Create: "/Users/rjeczalik/notify.tmp"
Hello from handler! (event=create, path=/Users/rjeczalik/notify.tmp)
...
```
```
~ $ cat > handler <<EOF
> echo "Hello from handler! (event={{.Event}}, path={{.Path}})"
> EOF

~ $ notify -f handler
2015/02/17 01:22:26 received notify.Create: "/Users/rjeczalik/notify.tmp"
Hello from handler! (event=create, path=/Users/rjeczalik/notify.tmp)
...
```

## cmd/prepend [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/cmd/prepend?status.png)](https://godoc.org/github.com/rjeczalik/tools/cmd/prepend)

Command prepend inserts data read from stdin or an input file at the
begining of the given file.

If data to prepend is passed both via stdin and input file, first the
given file is prepended with data read from stdin, then from input file.

The prepend command does not load the files contents to the memory,
making it suitable for large files. Writes issued by the prepend command
are atomic, meaning if reading from stdin or input file fails the
original file is left untouched.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/tools/cmd/prepend
```

*Documentation*

[godoc.org/github.com/rjeczalik/tools/cmd/prepend](http://godoc.org/github.com/rjeczalik/tools/cmd/prepend)

## netz [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/netz?status.png)](https://godoc.org/github.com/rjeczalik/tools/netz)

Package netz provides an interface for the `net` package from standard library.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/tools/netz
```

*Documentation*

[godoc.org/github.com/rjeczalik/tools/netz](http://godoc.org/github.com/rjeczalik/tools/netz)

## netz/memnetz [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/netz/memnetz?status.png)](https://godoc.org/github.com/rjeczalik/tools/netz/memnetz)

Package netz provides an implementation for an in-memory networking fake.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/tools/netz/memnetz
```

*Documentation*

[godoc.org/github.com/rjeczalik/tools/netz/memnetz](http://godoc.org/github.com/rjeczalik/tools/netz/memnetz)

