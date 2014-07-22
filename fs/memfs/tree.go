package memfs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Box drawings symbols - http://unicode-table.com/en/sections/box-drawing/.
var (
	boxVerticalRight = []byte("├")
	boxHorizontal    = []byte("─")
	boxVertical      = []byte("│")
	boxUpRight       = []byte("└")
	boxSpace         = []byte{'\u0020'}
	boxHardSpace     = []byte{'\u00A0'}
)

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

// CustomTree instructs tree builder how to parse single line of given buffer,
// where 'name' is the name of a tree node, 'depth' is its depth in the tree
// and 'err' eventual parsing failure. The 'line' is guaranteed to be non-nil
// non-empty.
type CustomTree func(line []byte) (depth int, name []byte, err error)

// Unix is a tree builder for the 'tree' Unix command.
var Unix CustomTree

func init() {
	Unix = func(p []byte) (depth int, name []byte, err error) {
		var n int
		// TODO(rjeczalik): Count up to first non-box character.
		depth = (bytes.Count(p, boxSpace) + bytes.Count(p, boxHardSpace) +
			bytes.Count(p, boxVertical)) / 4
		if n = bytes.LastIndex(p, boxHorizontal); n == -1 {
			err = fmt.Errorf("invalid syntax: %q", p)
			return
		}
		name = p[n:]
		if n = bytes.Index(name, boxSpace); n == -1 {
			err = fmt.Errorf("invalid syntax: %q", p)
			return
		}
		name = bytes.TrimSpace(name[n+1:])
		return
	}
}

// Create builds FS.Tree from given reader.
func (ct CustomTree) Create(r io.Reader) (fs FS, err error) {
	var (
		dir       = Directory{}
		buf       = bufio.NewReader(r)
		glob      []Directory
		name      []byte
		prevName  []byte
		depth     int
		prevDepth int
	)
	fs.Tree = dir
	line, err := buf.ReadBytes('\n')
	if len(line) == 0 || err == io.EOF {
		err = io.ErrUnexpectedEOF
		return
	}
	if err != nil {
		return
	}
	if len(line) != 1 || line[0] != '.' {
		p := filepath.FromSlash(string(bytes.TrimSpace(line)))
		if err = fs.MkdirAll(p, 0); err != nil {
			return
		}
		// TODO(rjeczalik): make it an exported helper method
		var perr *os.PathError
		if dir, perr = fs.lookup(p); perr != nil {
			err = perr
			return
		}
	}
	glob = append(glob, dir)
	// TODO(rjeczalik: handle empty directories (= names ending with '/')
	for {
		line, err = buf.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) == 0 {
			io.Copy(ioutil.Discard, buf)
			err, line = io.EOF, nil
		} else {
			depth, name, err = ct(line)
		}
		// Skip first iteration.
		if len(prevName) != 0 {
			// Insert the node from previous iteration - node is a directory when
			// a diference of the tree depth > 0, a file otherwise.
			p := string(prevName)
			switch {
			case depth > prevDepth:
				d := Directory{}
				dir[p], glob, dir = d, append(glob, dir), d
			case depth == prevDepth:
				dir[p] = File{}
			case depth < prevDepth:
				n := max(len(glob)+depth-prevDepth, 0)
				dir[p], dir, glob = File{}, glob[n], glob[:n]
			}
		}
		// A node from each iteration is handled on the next one. That's why the
		// error handling is deferred.
		if len(line) == 0 {
			if err == io.EOF {
				err = nil
			}
			return
		}
		prevDepth, prevName = depth, name
	}
}

// UnixTree builds FS.Tree from buffer that contains tree-like (Unix command) output.
//
// Example:
//
//   var tree = []byte(`.
//   └── dir
//       └── file.txt`)
//
//   fs, _ = memfs.FromTree(tree)
//   fmt.Printf("%#v\n", fs)
//   // Produces:
//   // memfs.FS{Tree: memfs.Directory{"dir": memfs.Directory{"file": memfs.File{}}}}
func UnixTree(p []byte) (FS, error) {
	return UnixTreeReader(bytes.NewBuffer(p))
}

// UnixTreeReader builds FS.Tree from io.Reader that contains tree-like output.
func UnixTreeReader(r io.Reader) (FS, error) {
	return Unix.Create(r)
}
