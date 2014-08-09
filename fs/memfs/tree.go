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

func dfs(d Directory, fn func(name string, item interface{}, state []dirQueue) error) (err error) {
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
		if err = fn(name, dq.Dir[s], glob); err != nil {
			return
		}
		if dir, ok := dq.Dir[s].(Directory); ok {
			if dirlen(dir) > 0 {
				glob = append(glob, newDirQueue(name, dir))
			}
		}
	}
	return
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

// EncodingState denotes a part of a tree, which is currently being printed by
// the TreeBuilder.Encode method.
type EncodingState uint8

// Example: Each Endcoding* variable makes the Unix.EncodeState function member
// return the corresponding bytes.
const (
	EncodingLevel     EncodingState = iota // returns []byte("│   ")
	EncodingLevelLast                      // returns []byte("    ")
	EncodingItem                           // returns []byte("├── ")
	EncodingItemLast                       // returns []byte("└── ")
)

// TreeBuilder provides an implementation of encoding.TextMarshaler and
// encoding.TextUnmarshaler for the memfs.FS struct. It may be configured to
// support custom formats by providing DecodeLine member function.
//
// Decoding
//
// DecodeLine instructs the TreeBuilder.Decode how to parse single line of given
// buffer, where 'name' is the name of a tree node, 'depth' is its depth in
// the tree and 'err' eventual parsing failure. The 'line' is guaranteed to be
// non-nil and non-empty.
//
// The function is expected to return non-nil and non-empty name and non-negative
// depth when err is nil. If the err is io.EOF, it will be translated to ErrTreeBuilder,
// because it will.
//
// If DecodeLine is nil, Unix.DecodeLine is used.
//
// Encoding
//
// EncodeState instructs the TreeBuilder.Encode how to print each part of the
// tree, denoted by a EncodingState argument passed to it. The output is not
// sanitized, which means e.g. newlines or other characters may break the tree
// layout.
//
// If EncodeState is nil, Unix.EncodeState is used.
type TreeBuilder struct {
	DecodeLine  func([]byte) (int, []byte, error)
	EncodeState func(EncodingState) []byte
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
	Unix.EncodeState = func(st EncodingState) []byte {
		// TODO(rjeczalik): map?
		switch st {
		case EncodingLevel:
			return boxDepth
		case EncodingItem:
			return boxItem
		case EncodingLevelLast:
			return boxDepthLast
		case EncodingItemLast:
			return boxItemLast
		}
		panic("unsupported encoding state")
	}
	Tab.DecodeLine = func(p []byte) (depth int, name []byte, err error) {
		depth = bytes.Count(p, []byte{'\t'})
		name = p[depth:]
		return
	}
	Tab.EncodeState = func(st EncodingState) []byte {
		switch st {
		case EncodingLevel, EncodingLevelLast, EncodingItem, EncodingItemLast:
			return []byte{'\t'}
		}
		panic("unsupported encoding state")
	}
}

// Decode builds fs.Tree from given reader using bt.DecodeLine callback for parsing
// node's name and its depth in the tree. Tree returns ErrTreeBuilder error when
// a call to ct gives invalid values.
//
// If tb.DecodeLine is nil, Unix.DecodeLine is used.
func (tb TreeBuilder) Decode(r io.Reader) (fs FS, err error) {
	var (
		e         error
		dir       = Directory{}
		buf       = bufio.NewReader(r)
		dec       = tb.DecodeLine
		glob      []Directory
		name      []byte
		prevName  []byte
		depth     int
		prevDepth int
	)
	fs.Tree = dir
	if dec == nil {
		dec = Unix.DecodeLine
	}
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

// Encode serializes the fs.Tree by writing its text representation to the w
// writer. The text is formatted according to the bt.EncodeState function.
//
// If tb.EncodeState is nil, Unix.EncodeState is used.
func (tb TreeBuilder) Encode(fs FS, w io.Writer) (err error) {
	var buf = bufio.NewWriter(w)
	defer func() {
		if err == nil {
			err = buf.Flush()
		}
	}()
	if dirlen(fs.Tree) == 0 {
		_, err = buf.WriteString(".\n")
		return
	}
	// TODO(rjeczalik): fold long root path
	if _, err = buf.WriteString(".\n"); err != nil {
		return
	}
	var enc = tb.EncodeState
	if enc == nil {
		enc = Unix.EncodeState
	}
	fn := func(s string, v interface{}, glob []dirQueue) (err error) {
		var dq = &glob[len(glob)-1]
		for i := 0; i < len(glob)-1; i++ {
			if len(glob[i].Queue) != 0 {
				if _, err = buf.Write(enc(EncodingLevel)); err != nil {
					return
				}
			} else {
				if _, err = buf.Write(enc(EncodingLevelLast)); err != nil {
					return
				}
			}
		}
		if len(dq.Queue) != 0 {
			if _, err = buf.Write(enc(EncodingItem)); err != nil {
				return
			}
		} else {
			if _, err = buf.Write(enc(EncodingItemLast)); err != nil {
				return
			}
		}
		if _, err = buf.WriteString(filepath.Base(s)); err != nil {
			return
		}
		if dir, ok := v.(Directory); ok && dirlen(dir) == 0 {
			if err = buf.WriteByte('/'); err != nil {
				return
			}
		}
		err = buf.WriteByte('\n')
		return
	}
	err = dfs(fs.Tree, fn)
	return
}

// MarshalUnix serializes FS.Tree to a text representation, which is the same as
// the Unix tree command's output.
//
// Example
//
//   var fs = memfs.FS{
//              Tree: memfs.Directory{
//                "dir": memfs.Directory{
//                  "file.txt": memfs.File{},
//                },
//              },
//            }
//
//   var tree, _ = memfs.MarshalUnix(fs)
//
// The tree slice is equal to:
//
//   var tree = []byte(`.
//   └── dir
//       └── file.txt`)
//
// MarshalUnix is a conveniance function which wraps Unix.Decode.
func MarshalUnix(fs FS) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, 128))
	if err := Unix.Encode(fs, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalUnix builds FS.Tree from a buffer that contains tree-like (Unix command) output.
//
// Example
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
// UnmarshalUnix(p) is a conveniance function which wraps Unix.Decode.
func UnmarshalUnix(p []byte) (FS, error) {
	return Unix.Decode(bytes.NewReader(p))
}

// MarshalTab serializes FS.Tree to a text representation, which is a \t-separated
// file tree.
//
// Example
//
//   var fs = memfs.FS{
//              Tree: memfs.Directory{
//                "dir": memfs.Directory{
//                  "file.txt": memfs.File{},
//                },
//              },
//            }
//
//   var tree, _ = memfs.MarshalTab(fs)
//
// The tree slice is equal to:
//
//   var tree = []byte(`.\ndir\n\tfile1.txt\n\tfile2.txt`)
//
// MarshalTab is a conveniance function which wraps Tab.Encode.
func MarshalTab(fs FS) ([]byte, error) {
	var buf bytes.Buffer
	if err := Tab.Encode(fs, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalTab builds FS.Tree from a buffer that contains \t-separated file tree.
//
// Example
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
// UnmarshalTab is a conveniance function which wraps Tab.Decode.
func UnmarshalTab(p []byte) (FS, error) {
	return Tab.Decode(bytes.NewReader(p))
}
