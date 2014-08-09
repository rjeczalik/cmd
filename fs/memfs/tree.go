package memfs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"unicode"
)

// TODO(rjeczalik): FS.String -> CustomPrinter type

// Box drawings symbols - http://unicode-table.com/en/sections/box-drawing/.
var (
	boxVerticalRight = []byte("├")
	boxHorizontal    = []byte("─")
	boxVertical      = []byte("│")
	boxUpRight       = []byte("└")
	boxSpace         = []byte{'\u0020'}
	boxHardSpace     = []byte{'\u00a0'}
)

var (
	boxDepth     = []byte("│\u00a0\u00a0\u0020")
	boxDepthLast = []byte("\u0020\u0020\u0020\u0020")
	boxItem      = []byte("├──\u0020")
	boxItemLast  = []byte("└──\u0020")
)

// String produces Unix-tree-like filesystem representation as a string.
//
// Example
//
// String can be use to convert between Tab and Unix tree representations, like
// in the following example:
//
//   var fs = memfs.Must(memfs.UnmarshalTab([]byte(".\ndir\n\tfile1.txt\n\tfile2.txt")))
//   fmt.Println(fs)
//
// Which prints:
//
//   .
//   └── dir
//       ├── file1.txt
//       └── file2.txt
func (fs FS) String() string {
	if dirlen(fs.Tree) == 0 {
		return ".\n"
	}
	var buf = bytes.NewBuffer(make([]byte, 0, 128))
	// TODO(rjeczalik): fold long root path
	buf.WriteByte('.')
	buf.WriteByte('\n')
	fn := func(s string, v interface{}, glob []dirQueue) bool {
		var dq = &glob[len(glob)-1]
		for i := 0; i < len(glob)-1; i++ {
			if len(glob[i].Queue) != 0 {
				buf.Write(boxDepth)
			} else {
				buf.Write(boxDepthLast)
			}
		}
		if len(dq.Queue) != 0 {
			buf.Write(boxItem)
		} else {
			buf.Write(boxItemLast)
		}
		buf.WriteString(filepath.Base(s))
		if dir, ok := v.(Directory); ok && dirlen(dir) == 0 {
			buf.WriteByte('/')
		}
		buf.WriteByte('\n')
		return true
	}
	dfs(fs.Tree, fn)
	return buf.String()
}

type dirQueue struct {
	Name  string
	Dir   Directory
	Queue []string
}

func newDirQueue(name string, dir Directory) dirQueue {
	return dirQueue{
		Name:  name,
		Dir:   dir,
		Queue: dir.Lsnames(OrderLexicalDesc),
	}
}

func dfs(d Directory, fn func(name string, item interface{}, state []dirQueue) bool) {
	if dirlen(d) == 0 {
		return
	}
	var glob = []dirQueue{newDirQueue("", d)}
	for len(glob) > 0 {
		var (
			s  string
			dq = &glob[len(glob)-1]
		)
		if len(dq.Queue) == 0 {
			glob = glob[:len(glob)-1]
			continue
		}
		s, dq.Queue = dq.Queue[len(dq.Queue)-1], dq.Queue[:len(dq.Queue)-1]
		name := filepath.Join(dq.Name, s)
		if !fn(name, dq.Dir[s], glob) {
			return
		}
		if dir, ok := dq.Dir[s].(Directory); ok {
			if dirlen(dir) > 0 {
				glob = append(glob, newDirQueue(name, dir))
			}
		}
	}
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

// ErrTreeBuilder represents a failure in handling returned values from
// a TreeBuilder.DecodeLine call.
var ErrTreeBuilder = errors.New("invalid name and/or depth values")

// TreeBuilder implements encoding.TextMarshaler and encoding.TextUnmarshaler
// for a FS structure. It may be configured to support custom formats by
// providing DecodeLine member function.
// DecodeLine instructs the TreeBuilder.Decode how to parse single line of given
// buffer, where 'name' is the name of a tree node, 'depth' is its depth in
// the tree and 'err' eventual parsing failure. The 'line' is guaranteed to be
// non-nil and non-empty.
// The function is expected to return non-nil and non-empty name and non-negative
// depth when err is nil. If the err is io.EOF, it will be translated to ErrTreeBuilder,
// because it will.
// If DecodeLine is nil, Tab.DecodeLine is used.
type TreeBuilder struct {
	DecodeLine func([]byte) (int, []byte, error)
}

// Unix is a tree builder for the 'tree' Unix command. It's guaranteed calls
// to Unix.Decode do not return ErrTreeBuilder.
var Unix TreeBuilder

// Tab is a tree builder for simplified tree representation, where each level
// is idented with one tabulation character (\t) only. It's guaranteed calls
// to Tab.Decode do not return ErrTreeBuilder.
var Tab TreeBuilder

func init() {
	Unix.DecodeLine = func(p []byte) (depth int, name []byte, err error) {
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
		name = name[n+1:]
		return
	}
	Tab.DecodeLine = func(p []byte) (depth int, name []byte, err error) {
		depth = bytes.Count(p, []byte{'\t'})
		name = p[depth:]
		return
	}
}

// Decode builds FS.Tree from given reader using ct.DecodeLine callback for parsing
// node's name and its depth in the tree. Tree returns ErrTreeBuilder error when
// a call to ct gives invalid values.
func (tb TreeBuilder) Decode(r io.Reader) (fs FS, err error) {
	var (
		e         error
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
		var perr *os.PathError
		if dir, perr = fs.lookup(p); perr != nil {
			err = perr
			return
		}
	}
	defer func() {
		// This may happen when ct failed to provide non-empty file name,
		// which left fs tree having a directory defined with a special key
		// which is not of Property type.
		if err == nil && !Fsck(fs) {
			err = errCorrupted
		}
	}()
	glob = append(glob, dir)
	for {
		line, err = buf.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) == 0 {
			// Drain the buffer, needed for some use-cases (encoding, net/rpc)
			io.Copy(ioutil.Discard, buf)
			err, line = io.EOF, nil
		} else {
			depth, name, e = tb.DecodeLine(bytes.TrimRightFunc(line, unicode.IsSpace))
			if len(name) == 0 || depth < 0 || e != nil {
				// Drain the buffer, needed for some use-cases (encoding, net/rpc)
				io.Copy(ioutil.Discard, buf)
				err, line = e, nil
				if err == nil || err == io.EOF {
					err = ErrTreeBuilder
				}
			}
		}
		// Skip first iteration.
		if len(prevName) != 0 {
			// Insert the node from previous iteration - node is a directory when
			// a diference of the tree depth > 0, a file otherwise.
			var (
				name  string
				value interface{}
			)
			if bytes.HasSuffix(prevName, []byte{'/'}) {
				name, value = string(bytes.TrimRight(prevName, "/")), Directory{}
			} else {
				name, value = string(prevName), File{}
			}
			switch {
			case depth > prevDepth:
				d := Directory{}
				dir[name], glob, dir = d, append(glob, dir), d
			case depth == prevDepth:
				dir[name] = value
			case depth < prevDepth:
				n := max(len(glob)+depth-prevDepth, 0)
				dir[name], dir, glob = value, glob[n], glob[:n]
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

// UnmarshalUnix builds FS.Tree from a buffer that contains tree-like (Unix command) output.
//
// Example:
//
//   var tree = []byte(`.
//   └── dir
//       └── file.txt`)
//
//   var fs = memfs.Must(memfs.UnmarshalUnix(tree))
//
// The above is an equivalent to:
//
//   var fs = memfs.FS{
//              Tree: memfs.Directory{
//                "dir": memfs.Directory{
//                  "file.txt": memfs.File{},
//                },
//              },
//            }
//
// UnmarshalUnix(p) is a short alternative to the Unix.Decode(bytes.NewReader(p)).
func UnmarshalUnix(p []byte) (FS, error) {
	return Unix.Decode(bytes.NewReader(p))
}

// UnmarshalTab builds FS.Tree from a buffer that contains \t-separated file tree.
//
// Example:
//
//   var tree = []byte(`.\ndir\n\tfile1.txt\n\tfile2.txt`)
//   var fs = memfs.Must(memfs.UnmarshalTab(tree))
//
// The above is an equivalent to:
//
//   var fs = memfs.FS{
//              Tree: memfs.Directory{
//                "dir": memfs.Directory{
//                  "file1.txt": memfs.File{},
//                  "file2.txt": memfs.File{},
//                },
//              },
//            }
//
// UnmarshalTab(p) is a short alternative to the Tab.Decode(bytes.NewReader(p)).
func UnmarshalTab(p []byte) (FS, error) {
	return Tab.Decode(bytes.NewReader(p))
}
