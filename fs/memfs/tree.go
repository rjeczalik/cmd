package memfs

import (
	"bufio"
	"bytes"
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

// FromTree builds FS.Tree from buffer that contains tree-like (Unix command) output.
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
func FromTree(p []byte) (FS, error) {
	return FromTreeReader(bytes.NewBuffer(p))
}

// FromTreeReader builds FS.Tree from io.Reader that contains tree-like output.
func FromTreeReader(r io.Reader) (fs FS, err error) {
	var (
		dir       = Directory{}
		buf       = bufio.NewReader(r)
		glob      []Directory
		name      []byte
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
		}
		// Estimate depth. Hacky way to avoid context parsing.
		depth := (bytes.Count(line, boxSpace) + bytes.Count(line, boxHardSpace) +
			bytes.Count(line, boxVertical)) / 4
		// Skip first iteration.
		if len(name) != 0 {
			// Insert the node from previous iteration - node is a directory when
			// a diference of the tree depth > 0, a file otherwise.
			p := string(name)
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
				break
			}
			if err != nil {
				return
			}
		}
		// Parse a name of the the current node.
		name = line[bytes.LastIndex(line, boxHorizontal):]
		name = bytes.TrimSpace(name[bytes.Index(name, boxSpace)+1:])
		prevDepth = depth
	}
	return
}
