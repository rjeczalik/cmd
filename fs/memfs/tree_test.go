package memfs

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

var cases = [...]FS{
	0: {
		Tree: Directory{
			"dir": Directory{
				"file.txt": File{},
			},
		},
	},
	1: {
		Tree: Directory{
			"out":         Directory{},
			"out.gif":     File{},
			"out.ogv":     File{},
			"output2.gif": File{},
			"output3.gif": File{},
			"output.gif":  File{},
		},
	},
	2: {
		Tree: Directory{
			"github.com": Directory{
				"rjeczalik": Directory{
					"tools": Directory{
						"doc.go": File{},
						"fs": Directory{
							"fs.go": File{},
							"glob": Directory{
								"glob.go":      File{},
								"glob_test.go": File{},
							},
							"memfs": Directory{
								"memfs.go":      File{},
								"memfs_test.go": File{},
								"tree.go":       File{},
								"tree_test.go":  File{},
								"util.go":       File{},
								"util_test.go":  File{},
							},
							"testdata": Directory{
								"test":     Directory{},
								"tree.txt": File{},
							},
						},
						"LICENSE": File{},
						"netz": Directory{
							"memnetz": Directory{
								"memnetz.go":      File{},
								"memnetz_test.go": File{},
							},
							"netz.go":       File{},
							"split.go":      File{},
							"split_test.go": File{},
						},
						"README.md": File{},
						"tmp":       Directory{},
					},
				},
			},
		},
	},
	3: {
		Tree: Directory{
			"a": Directory{
				"b1": Directory{
					"c1": Directory{
						"c1.txt": File{},
					},
					"c2": Directory{
						"c2.txt": File{},
					},
					"c3": Directory{
						"c3.txt": File{},
						"d1": Directory{
							"e1": Directory{
								"_": Directory{
									"_.txt": File{},
								},
								"e1.txt": File{},
								"e2.txt": File{},
								"e":      Directory{},
							},
						},
					},
				},
				"b2": Directory{
					"c1": Directory{
						"d1.txt": File{},
						"d2":     Directory{},
						"d3.txt": File{},
					},
				},
			},
			"a.txt": File{},
			"w": Directory{
				"w.txt": File{},
				"x": Directory{
					"y": Directory{
						"z": Directory{
							"1.txt": File{},
						},
					},
					"y.txt": File{},
				},
			},
		},
	},
}

var unix = [...][]byte{
	0: []byte(".\n└── dir\n    ├── file.txt"),
	1: []byte(`.
├── out/
├── out.gif
├── out.ogv
├── output2.gif
├── output3.gif
└── output.gif`),
	2: []byte(`/github.com/rjeczalik/tools
├── doc.go
├── fs
│   ├── fs.go
│   ├── glob
│   │   ├── glob.go
│   │   └── glob_test.go
│   ├── memfs
│   │   ├── memfs.go
│   │   ├── memfs_test.go
│   │   ├── tree.go
│   │   ├── tree_test.go
│   │   ├── util.go
│   │   └── util_test.go
│   └── testdata
│       ├── test/
│       └── tree.txt
├── LICENSE
├── netz
│   ├── memnetz
│   │   ├── memnetz.go
│   │   └── memnetz_test.go
│   ├── netz.go
│   ├── split.go
│   └── split_test.go
├── README.md
└── tmp/`),
	3: []byte(`.
├── a
│   ├── b1
│   │   ├── c1
│   │   │   └── c1.txt
│   │   ├── c2
│   │   │   └── c2.txt
│   │   └── c3
│   │       ├── c3.txt
│   │       └── d1
│   │           └── e1
│   │               ├── _
│   │               │   └── _.txt
│   │               ├── e1.txt
│   │               ├── e2.txt
│   │               └── e/
│   └── b2
│       └── c1
│           ├── d1.txt
│           ├── d2/
│           └── d3.txt
├── a.txt
└── w
    ├── w.txt
    └── x
        ├── y
        │   └── z
        │       └── 1.txt
        └── y.txt

16 directories, 12 files`),
}

var tab = [...][]byte{
	0: []byte(".\ndir\n\tfile.txt"),
	1: []byte(`.
out/
out.gif
out.ogv
output2.gif
output3.gif
output.gif`),
	2: []byte(`/github.com/rjeczalik/tools
doc.go
fs
	fs.go
	glob
		glob.go
		glob_test.go
	memfs
		memfs.go
		memfs_test.go
		tree.go
		tree_test.go
		util.go
		util_test.go
	testdata
		test/
		tree.txt
LICENSE
netz
	memnetz
		memnetz.go
		memnetz_test.go
	netz.go
	split.go
	split_test.go
README.md
tmp/`),
	3: []byte(`.
a
	b1
		c1
			c1.txt
		c2
			c2.txt
		c3
			c3.txt
			d1
				e1
					_
						_.txt
					e1.txt
					e2.txt
					e/
	b2
		c1
			d1.txt
			d2/
			d3.txt
a.txt
w
	w.txt
	x
		y
			z
				1.txt
		y.txt`),
}

func TestString(t *testing.T) {
	for i, cas := range cases {
		fs, err := Unix.Decode(strings.NewReader(cas.String()))
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if !Equal(fs, cas) {
			t.Errorf("want Compare(...)=true; got false (i=%d)", i)
		}
	}
}

func TestUnixTree(t *testing.T) {
	for i, p := range unix {
		fs, err := UnmarshalUnix(p)
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if !Equal(fs, cases[i]) {
			t.Errorf("want Compare(...)=true; got false (i=%d)", i)
		}
	}
}

func TestTabTree(t *testing.T) {
	for i, p := range tab {
		fs, err := UnmarshalTab(p)
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if !Equal(fs, cases[i]) {
			t.Errorf("want Compare(...)=true; got false (i=%d)", i)
		}
	}
}

func TestCustomTreeErr(t *testing.T) {
	cases := [...][]struct {
		depth int
		name  string
		err   error
	}{
		0: {
			{0, "a", nil},
			{1, "b", nil},
			{1, "", nil},
			{0, "d", nil},
		},
		1: {
			{0, "a", nil},
			{0, "b", nil},
			{-1, "c", nil},
			{0, "d", nil},
		},
		2: {
			{0, "a", nil},
			{1, "b", nil},
			{2, "c", ErrTreeBuilder},
			{3, "d", nil},
		},
	}

	p := make([]byte, 8)
	for i, cas := range cases {
		j := 0
		tb := TreeBuilder{DecodeLine: func([]byte) (n int, p []byte, e error) {
			n, p, e = cas[j].depth, []byte(cas[j].name), cas[j].err
			if j < len(cas)-1 {
				j++
			}
			return
		}}
		buf := bytes.NewReader(tab[1])
		if _, err := tb.Decode(buf); err == nil {
			t.Errorf("want err to be non-nil (i=%d)", i)
			continue
		}
		if _, err := buf.Read(p); err != io.EOF {
			t.Errorf("want err=io.EOF; got %q (i=%d)", err, i)
		}
	}
}
