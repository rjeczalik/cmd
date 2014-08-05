package fsutil

import (
	"path/filepath"
	"testing"

	fs "github.com/rjeczalik/tools/fs/memfs"
)

var testdata = fs.Must(fs.TabTree([]byte(`.
data
	github.com
		user
			example
				.git/
				dir
					dir.txt
				assets
					js
						app.js
						link.js
					css
						default.css
src
	github.com
		user
			example
				.git/
				dir
					dir.go
				example.go`)))

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

func TestReadpaths(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

func TestReaddirpaths(t *testing.T) {
	cases := map[string][]string{
		filepath.FromSlash("/data/github.com/user/example"): {
			"assets",
			"dir",
		},
		filepath.FromSlash("/src/github.com/user/example"): {
			"dir",
		},
	}
	c := Control{FS: testdata}
	for dir, cas := range cases {
		for _, b := range [...]bool{false, true} {
			if c.Hidden = b; b {
				cas = append(cas, ".git")
			}
			names := c.Readdirpaths(dir)
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
	g := Control{FS: testdata}
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

func TestFind(t *testing.T) {
	t.Skip("TODO(rjeczalik)")
}

var schema = fs.FS{
	Tree: fs.Directory{
		"schema": fs.Directory{
			"licstat": fs.Directory{
				"schema": fs.Directory{
					"databasequery": fs.Directory{
						"reqaddaliasls.json":  fs.File{},
						"reqdeletef.json":     fs.File{},
						"reqdeletels.json":    fs.File{},
						"reqmergels.json":     fs.File{},
						"reqquerystatus.json": fs.File{},
					},
					"generalinfo": fs.Directory{
						"reqinstallpath.json": fs.File{},
					},
					"license": fs.Directory{
						"reqlicensedetail.json": fs.File{},
					},
					"monitorconf": fs.Directory{
						"reqaddls.json":    fs.File{},
						"reqcheckls.json":  fs.File{},
						"reqeditls.json":   fs.File{},
						"reqremovels.json": fs.File{},
						"reqstatusls.json": fs.File{},
					},
					"definitions.json": fs.File{},
				},
			},
		},
		"src": fs.Directory{
			"licstat": fs.Directory{
				"schema": fs.Directory{
					"tmp":       fs.Directory{},
					"schema.go": fs.File{},
				},
			},
		},
	},
}

func TestIntersect_SchemaUnique(t *testing.T) {
	cas := []string{
		filepath.FromSlash("licstat/schema"),
	}
	names := (Control{FS: schema}).Intersect(filepath.FromSlash("/src"), filepath.FromSlash("/schema"))
	if names == nil {
		t.Fatal("want names!=nil")
	}
	if !equal(names, cas) {
		t.Errorf("want names=%v; got %v", cas, names)
	}
}
