package glob

import (
	"path/filepath"
	"testing"

	"github.com/rjeczalik/tools/fs/fakefs"
)

var testdata = fakefs.FS{
	Tree: fakefs.Directory{
		"data": fakefs.Directory{
			"github.com": fakefs.Directory{
				"user": fakefs.Directory{
					"example": fakefs.Directory{
						".git": fakefs.Directory{},
						"assets": fakefs.Directory{
							"js": fakefs.Directory{
								"app.js":  fakefs.NewFile("app.js"),
								"link.js": fakefs.NewFile("link.js"),
							},
							"css": fakefs.Directory{
								"default.css": fakefs.NewFile("default.css"),
							},
						},
						"dir": fakefs.Directory{
							"dir.txt": fakefs.NewFile("dir.txt"),
						},
					},
				},
			},
		},
		"src": fakefs.Directory{
			"github.com": fakefs.Directory{
				"user": fakefs.Directory{
					"example": fakefs.Directory{
						".git": fakefs.Directory{},
						"dir": fakefs.Directory{
							"dir.go": fakefs.NewFile("dir.go"),
						},
						"example.go": fakefs.NewFile("example.go"),
					},
				},
			},
		},
	},
}

func equal(lhs, rhs []string) bool {
	if len(lhs) != len(rhs) {
		return false
	}
LOOP:
	for i := range lhs {
		for j := range rhs {
			if lhs[i] == rhs[j] {
				continue LOOP
			}
		}
		return false
	}
	return true
}

func TestReaddirnames(t *testing.T) {
	cases := map[string][]string{
		filepath.FromSlash("/data/github.com/user/example"): {
			"assets",
			"dir",
		},
		filepath.FromSlash("/src/github.com/user/example"): {
			"dir",
		},
	}
	g := Glob{FS: testdata}
	for dir, cas := range cases {
		for _, b := range [...]bool{false, true} {
			if g.Hidden = b; b {
				cas = append(cas, ".git")
			}
			names := g.Readdirnames(dir)
			if names == nil {
				t.Errorf("want names!=nil (dir=%q,hidden=%v)", dir, b)
				continue
			}
			if !equal(names, cas) {
				t.Errorf("want names=%v; got %v (dir=%q,hidden=%v)", cas, names, dir, b)
			}
		}
	}
}

func TestIntersect(t *testing.T) {
	cas := []string{
		filepath.FromSlash("github.com/user/example"),
		filepath.FromSlash("github.com/user/example/dir"),
	}
	g := Glob{FS: testdata}
	for _, b := range [...]bool{false, true} {
		if g.Hidden = b; b {
			cas = append(cas, filepath.FromSlash("github.com/user/example/.git"))
		}
		names := g.Intersect(filepath.FromSlash("/src"), filepath.FromSlash("/data"))
		if names == nil {
			t.Errorf("want names!=nil (hidden=%v)", b)
			continue
		}
		if !equal(names, cas) {
			t.Errorf("want names=%v; got %v (hidden=%v)", cas, names, b)
		}
	}
}
