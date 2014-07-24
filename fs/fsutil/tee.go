package fsutil

import (
	"os"
	"path/filepath"

	"github.com/rjeczalik/tools/fs"
)

type teefs struct {
	fs.Filesystem
	mirror fs.Filesystem
}

func (tf teefs) Create(path string) (f fs.File, err error) {
	if f, err = tf.Filesystem.Create(path); err != nil {
		return
	}
	var fi os.FileInfo
	if fi, err = f.Stat(); err != nil {
		f.Close()
		return
	}
	if err = tf.mirror.MkdirAll(filepath.Base(path), fi.Mode()); err != nil {
		f.Close()
		return
	}
	var mf fs.File
	if mf, err = tf.mirror.Create(path); err != nil {
		f.Close()
		return
	}
	mf.Close()
	return
}

func (tf teefs) Mkdir(path string, perm os.FileMode) error {
	if err := tf.Filesystem.Mkdir(path, perm); err != nil {
		return err
	}
	return tf.mirror.MkdirAll(path, perm)
}

func (tf teefs) MkdirAll(path string, perm os.FileMode) (err error) {
	if err := tf.Filesystem.MkdirAll(path, perm); err != nil {
		return err
	}
	return tf.mirror.MkdirAll(path, perm)
}

func (tf teefs) Open(path string) (f fs.File, err error) {
	if f, err = tf.Filesystem.Open(path); err != nil {
		return
	}
	var fi os.FileInfo
	if fi, err = f.Stat(); err != nil {
		f.Close()
		return
	}
	if fi.IsDir() {
		if err = tf.mirror.MkdirAll(path, fi.Mode()); err == nil {
			f = teefile{
				File:   f,
				dir:    path,
				mirror: tf.mirror,
			}
		}
	} else {
		if err = tf.mirror.MkdirAll(filepath.Dir(path), fi.Mode()); err != nil {
			f.Close()
			return
		}
		var mf fs.File
		if fi, err = tf.mirror.Stat(path); os.IsNotExist(err) {
			if mf, err = tf.mirror.Create(path); err == nil {
				mf.Close()
			}
		}
	}
	if err != nil {
		f.Close()
	}
	return
}

type teefile struct {
	fs.File
	dir    string
	mirror fs.Filesystem
}

func (tf teefile) Readdir(n int) (fi []os.FileInfo, err error) {
	if fi, err = tf.File.Readdir(n); err != nil {
		return
	}
	for _, fi := range fi {
		path := filepath.Join(tf.dir, filepath.Base(fi.Name()))
		if fi.IsDir() {
			err = tf.mirror.Mkdir(path, fi.Mode())
		} else {
			var f fs.File
			if f, err = tf.mirror.Create(path); err == nil {
				f.Close()
			}
		}
	}
	return
}

// TeeFilesystem returns a filesystem which mirrors file tree read from 'read'
// filesystem in the 'write' one. Every path passed to the Open, Create, Mkdir,
// MkdirAll and Readdir methods is created on the 'write' filesystem only if
// the call was sucessful.
// TeeFilesystem can be used as a spy for recording file and/or directory access
// of the 'read' filesystem.
func TeeFilesystem(read, write fs.Filesystem) fs.Filesystem {
	return teefs{
		Filesystem: read,
		mirror:     write,
	}
}
