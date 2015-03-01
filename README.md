tools [![Build Status](https://img.shields.io/travis/rjeczalik/tools/master.svg)](https://travis-ci.org/rjeczalik/tools "linux_amd64") [![Build Status](https://img.shields.io/travis/rjeczalik/tools/osx.svg)](https://travis-ci.org/rjeczalik/tools "darwin_amd64") [![Build status](https://img.shields.io/appveyor/ci/rjeczalik/tools-161.svg)](https://ci.appveyor.com/project/rjeczalik/tools-161 "windows_amd64") [![Coverage Status](https://img.shields.io/coveralls/rjeczalik/tools/master.svg)](https://coveralls.io/r/rjeczalik/tools?branch=master)
=====

Handmade tools for day-to-day plumbing.

*Installation*

```
~ $ go get -u github.com/rjeczalik/tools/cmd/...
```

## cmd/notify [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/cmd/notify?status.png)](https://godoc.org/github.com/rjeczalik/tools/cmd/notify)

Listens on filesystem changes and forwards received mapping to user-defined handlers.

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

## cmd/dln [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/cmd/dln?status.png)](https://godoc.org/github.com/rjeczalik/tools/cmd/dln)

Prints derivative computed out of line-separated numbers.

*Documentation*

[godoc.org/github.com/rjeczalik/tools/cmd/dln](http://godoc.org/github.com/rjeczalik/tools/cmd/dln)

*Usage*

```
~ $ curl -sS http://cdimage.ubuntu.com/daily-live/current/vivid-desktop-amd64.iso -o vivid-amd64.iso &
[1] 21496
```
```
~ $ while sleep 1; do du -BK vivid-amd64.iso ; done | dln
1008
1172
1548
2332
3200
4756
6572
7056
7052
7048
7060
7036
7048
7056
^C
```

## cmd/hist [![GoDoc](https://godoc.org/github.com/rjeczalik/tools/cmd/hist?status.png)](https://godoc.org/github.com/rjeczalik/tools/cmd/hist)

Prints histogram for line-separated data points. It sorts the result set by the number of occurances in descending order.

*Documentation*

[godoc.org/github.com/rjeczalik/tools/cmd/hist](http://godoc.org/github.com/rjeczalik/tools/cmd/hist)

*Usage*

```
~ $ log=https://gist.githubusercontent.com/rjeczalik/f18349ad629f07d19839/raw/b8089282fdd5a8ea8589fe33bc88cc6d29db7026/lazyvm.log
```
```
~ $ curl -sS $log | dln | hist
  0	962	░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
  1	5	
 18	3	
  3	2	
  9	1	
  6	1	
 49	1	
 23	1	
 22	1	
 21	1	
 59	1	
 11	1	
 78	1	
  4	1	
 42	1	
  2	1
```
```
~ $ curl -sS $log | dln | hist -slice 1:
  1	5	░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
 18	3	░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
  3	2	░░░░░░░░░░░░░░░░░░░░░░░░░
  9	1	░░░░░░░░░░░░
  6	1	░░░░░░░░░░░░
 49	1	░░░░░░░░░░░░
 23	1	░░░░░░░░░░░░
 22	1	░░░░░░░░░░░░
 21	1	░░░░░░░░░░░░
 59	1	░░░░░░░░░░░░
 11	1	░░░░░░░░░░░░
 78	1	░░░░░░░░░░░░
  4	1	░░░░░░░░░░░░
 42	1	░░░░░░░░░░░░
  2	1	░░░░░░░░░░░░
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

*Documentation*

[godoc.org/github.com/rjeczalik/tools/cmd/prepend](http://godoc.org/github.com/rjeczalik/tools/cmd/prepend)

*Usage*

```
~ $ cat > preamble.txt <<EOF
> // Copyright (c) 2015 Your Team. All right reserved.
> // Use of this source code is governed by the X license
> // that can be found in the LICENSE file
>
> EOF

~ $ preprend -u -f preamble.txt *.go
```

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

