package memfs

import "testing"

var cases = [...]struct {
	p  []byte
	fs FS
}{{
	[]byte(`.
├── out.gif
├── out.ogv
├── output2.gif
├── output3.gif
└── output.gif`),
	FS{
		Tree: Directory{
			"out.gif":     File{},
			"out.ogv":     File{},
			"output2.gif": File{},
			"output3.gif": File{},
			"output.gif":  File{},
		},
	},
}, {
	[]byte(`/github.com/rjeczalik/tools
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
│       └── tree.txt
├── LICENSE
├── netz
│   ├── memnetz
│   │   ├── memnetz.go
│   │   └── memnetz_test.go
│   ├── netz.go
│   ├── split.go
│   └── split_test.go
└── README.md`),
	FS{
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
					},
				},
			},
		},
	},
}, {
	[]byte(`.
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
│   │               └── e2.txt
│   └── b2
│       └── c1
│           ├── d1.txt
│           ├── d2.txt
│           └── d3.txt
├── a.txt
└── w
    ├── w.txt
    └── x
        ├── y
        │   └── z
        │       └── 1.txt
        └── y.txt

14 directories, 13 files`),
	FS{
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
							},
						},
					},
				},
				"b2": Directory{
					"c1": Directory{
						"d1.txt": File{},
						"d2.txt": File{},
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
}}

func TestUnixTree(t *testing.T) {
	for i, cas := range cases {
		fs, err := UnixTree(cas.p)
		if err != nil {
			t.Errorf("want err=nil; got %q (i=%d)", err, i)
			continue
		}
		if !Compare(&fs, &cas.fs) {
			t.Errorf("want Compare(...)=true; got false (i=%d)", i)
		}
	}
}
