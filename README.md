tools [![Build Status](https://travis-ci.org/rjeczalik/tools.png?branch=master)](https://travis-ci.org/rjeczalik/tools "linux_amd64") [![Build Status](https://travis-ci.org/rjeczalik/tools.png?branch=osx)](https://travis-ci.org/rjeczalik/tools "darwin_amd64") [![Build status](https://ci.appveyor.com/api/projects/status/2p6hms7nuantflwb)](https://ci.appveyor.com/project/rjeczalik/tools-161 "windows_amd64")
=====

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
github.com
└── rjeczalik
    └── tools
        └── fs
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

6 directories, 11 files
```

**NOTE** `fs.Filesystem` does not support symlinks yet ([#3](https://github.com/rjeczalik/tools/issues/3)), that's why `gotree` will print any symlink as regular file or directory. Moreover it won't follow nor resolve any of them.

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
~ $ tree
.
├── dir
│   └── file.txt
└── file.txt

1 directory, 2 files
~ $ tree | mktree -o /tmp/mktree
~ $ tree /tmp/mktree
/tmp/mktree
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

