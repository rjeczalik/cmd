tools [![Build Status](https://img.shields.io/travis/rjeczalik/tools/master.svg)](https://travis-ci.org/rjeczalik/tools "linux_amd64") [![Build Status](https://img.shields.io/travis/rjeczalik/tools/osx.svg)](https://travis-ci.org/rjeczalik/tools "darwin_amd64") [![Build status](https://img.shields.io/appveyor/ci/rjeczalik/tools-161.svg)](https://ci.appveyor.com/project/rjeczalik/tools-161 "windows_amd64") [![Coverage Status](https://img.shields.io/coveralls/rjeczalik/tools/master.svg)](https://coveralls.io/r/rjeczalik/tools?branch=master)
=====

Productivity commands. Interfaces and mocks for os and net packages. 

* Packages
  * [netz](README.md#netz-)
  * [netz/memnetz](README.md#netzmemnetz-)
  * [rw](README.md#rw-)

## rw [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/rw?status.png)](https://godoc.org/github.com/rjeczalik/tools/rw)

Package rw provides various utilities implementing wrappers for io.Reader and io.Writer.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/tools/rw
```

*Documentation*

[godoc.org/github.com/rjeczalik/tools/rw](http://godoc.org/github.com/rjeczalik/tools/rw)

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

