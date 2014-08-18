tools [![Build Status](https://img.shields.io/travis/rjeczalik/tools/master.svg)](https://travis-ci.org/rjeczalik/tools "linux_amd64") [![Build Status](https://img.shields.io/travis/rjeczalik/tools/osx.svg)](https://travis-ci.org/rjeczalik/tools "darwin_amd64") [![Build status](https://img.shields.io/appveyor/ci/rjeczalik/tools-161.svg)](https://ci.appveyor.com/project/rjeczalik/tools-161 "windows_amd64") [![Coverage Status](https://img.shields.io/coveralls/rjeczalik/tools/master.svg)](https://coveralls.io/r/rjeczalik/tools?branch=master)
=====

Productivity commands. Interfaces and mocks for os and net packages. 

* Commands
  * [cmd/gotree](README.md#cmdgotree-)
  * [cmd/mktree](README.md#cmdmktree-)

* Packages
  * [fs](README.md#fs-)
  * [fs/memfs](README.md#fsmemfs-)
  * [fs/fsutil](README.md#fsfsutil-)
  * [netz](README.md#netz-)
  * [netz/memnetz](README.md#netzmemnetz-)
  * [rw](README.md#rw-)

## cmd/gotree [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/cmd/gotree?status.png)](https://godoc.org/github.com/rjeczalik/tools/cmd/gotree)

Command `gotree` is Go implementation of the Unix `tree` command.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/tools/cmd/gotree
```

*Documentation*

[godoc.org/github.com/rjeczalik/tools/cmd/gotree](http://godoc.org/github.com/rjeczalik/tools/cmd/gotree)

*Usage*

```bash
~/src $ gotree github.com/rjeczalik/tools/fs
github.com/rjeczalik/tools/fs/.
├── fs.go
├── fsutil
│   ├── fsutil.go
│   ├── fsutil_test.go
│   ├── tee.go
│   └── tee_test.go
└── memfs
    ├── memfs.go
    ├── memfs_test.go
    ├── tree.go
    ├── tree_test.go
    ├── util.go
    └── util_test.go

2 directories, 11 files
```

**NOTE** `fs.Filesystem` does not support symlinks yet ([#3](https://github.com/rjeczalik/tools/issues/3)), that's why `gotree` will print any symlink as regular file or directory. Moreover it won't follow nor resolve any of them.

```bash
~/src $ gotree -go=80 github.com/rjeczalik/tools/fs
memfs.Must(memfs.UnmarshalTab([]byte(".\n\tfs.go\n\tfsutil\n\t\tfsutil.go" +
	"\n\t\tfsutil_test.go\n\t\ttee.go\n\t\ttee_test.go\n\tmemfs\n\t\tmem" +
	"fs.go\n\t\tmemfs_test.go\n\t\ttree.go\n\t\ttree_test.go\n\t\tutil.g" +
	"o\n\t\tutil_test.go\n")))
```
```bash
~/src $ gotree -var=fspkg github.com/rjeczalik/tools/fs
var fspkg = memfs.Must(memfs.UnmarshalTab([]byte(".\n\tfs.go\n\tfsutil\n\t" +
	"\tfsutil.go\n\t\tfsutil_test.go\n\t\ttee.go\n\t\ttee_test.go\n\tmem" +
	"fs\n\t\tmemfs.go\n\t\tmemfs_test.go\n\t\ttree.go\n\t\ttree_test.go\n" +
	"\t\tutil.go\n\t\tutil_test.go\n")))
```

## cmd/mktree [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/cmd/mktree?status.png)](https://godoc.org/github.com/rjeczalik/tools/cmd/mktree)

Command mktree creates a file tree out of `tree` output read from standard input.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/tools/cmd/mktree
```

*Documentation*

[godoc.org/github.com/rjeczalik/tools/cmd/mktree](http://godoc.org/github.com/rjeczalik/tools/cmd/mktree)

*Usage*

```bash
~ $ gotree
.
├── dir
│   └── file.txt
└── file.txt

1 directory, 2 files

~ $ gotree | mktree -o /tmp/mktree

~ $ gotree /tmp/mktree
/tmp/mktree
├── dir
│   └── file.txt
└── file.txt

1 directory, 2 files
```
```bash
~ $ gotree > tree.txt

~ $ mktree -o /tmp/mktree2 tree.txt

~ $ gotree /tmp/mktree2
/tmp/mktree2
├── dir
│   └── file.txt
└── file.txt

1 directory, 2 files
```

## fs [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/fs?status.png)](https://godoc.org/github.com/rjeczalik/tools/fs)

Package fs provides an interface for the filesystem-related functions from the `os` package.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/tools/fs
```

*Documentation*

[godoc.org/github.com/rjeczalik/tools/fs](http://godoc.org/github.com/rjeczalik/tools/fs)

## fs/memfs [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/fs/memfs?status.png)](https://godoc.org/github.com/rjeczalik/tools/fs/memfs)

Package memfs provides an implementation for an in-memory filesystem.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/tools/fs/memfs
```

*Documentation*

[godoc.org/github.com/rjeczalik/tools/fs/memfs](http://godoc.org/github.com/rjeczalik/tools/fs/memfs)

## fs/fsutil [![GoDoc](https://godoc.org/github.com/rjeczalik/fs/tools/fsutil?status.png)](https://godoc.org/github.com/rjeczalik/tools/fs/fsutil)

Package fsutil is a collection of various filesystem utility functions.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/tools/fs/fsutil
```

*Documentation*

[godoc.org/github.com/rjeczalik/tools/fs/fsutil](http://godoc.org/github.com/rjeczalik/tools/fs/fsutil)

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

