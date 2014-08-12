package fsutil

import (
	"os"
	"path/filepath"

	"github.com/rjeczalik/tools/fs"
)

type teefs struct {
	read  fs.Filesystem
	write fs.Filesystem
}

func (tf teefs) writefi(path string, fi os.FileInfo) (err error) {
	if fi.IsDir() {
		err = tf.write.MkdirAll(path, fi.Mode())
	} else {
		if err = tf.write.MkdirAll(filepath.Dir(path), fi.Mode()); err != nil {
			return
		}
		if fi, err = tf.write.Stat(path); os.IsNotExist(err) {
			var mf fs.File
			if mf, err = tf.write.Create(path); err == nil {
				mf.Close()
			}
		}
	}
	return
}

func (tf teefs) Create(path string) (f fs.File, err error) {
	if f, err = tf.read.Create(path); err != nil {
		return
	}
	var fi os.FileInfo
	if fi, err = f.Stat(); err != nil {
		f.Close()
		return
	}
	if err = tf.write.MkdirAll(filepath.Base(path), fi.Mode()); err != nil {
		f.Close()
		return
	}
	var mf fs.File
	if mf, err = tf.write.Create(path); err != nil {
		f.Close()
		return
	}
	mf.Close()
	return
}

func (tf teefs) Mkdir(path string, perm os.FileMode) error {
	if err := tf.read.Mkdir(path, perm); err != nil {
		return err
	}
	return tf.write.MkdirAll(path, perm)
}

func (tf teefs) MkdirAll(path string, perm os.FileMode) (err error) {
	if err := tf.read.MkdirAll(path, perm); err != nil {
		return err
	}
	return tf.write.MkdirAll(path, perm)
}

func (tf teefs) Open(path string) (f fs.File, err error) {
	if f, err = tf.read.Open(path); err != nil {
		return
	}
	var fi os.FileInfo
	if fi, err = f.Stat(); err != nil {
		f.Close()
		return nil, err
	}
	if err = tf.writefi(path, fi); err != nil {
		f.Close()
		return nil, err
	}
	if fi.IsDir() {
		f = teefile{
			File:  f,
			dir:   path,
			write: tf.write,
		}
	}
	return
}

func (tf teefs) Remove(path string) (err error) {
	if err = tf.read.Remove(path); err != nil {
		return
	}
	// Ignore error in case the path weren't recorded before by Open or Create
	// methods.
	tf.write.Remove(path)
	return
}

func (tf teefs) RemoveAll(path string) (err error) {
	if err = tf.read.RemoveAll(path); err != nil {
		return
	}
	// Ignore error in case the path weren't recorded before by Open or Create
	// methods.
	tf.write.RemoveAll(path)
	return

}

func (tf teefs) Stat(path string) (fi os.FileInfo, err error) {
	if fi, err = tf.read.Stat(path); err != nil {
		return
	}
	if err = tf.writefi(path, fi); err != nil {
		return nil, err
	}
	return
}

func (tf teefs) Walk(path string, fn filepath.WalkFunc) error {
	// TODO(rjeczalik): call fn on tf.write or writefi after every fn?
	return tf.read.Walk(path, fn)
}

type teefile struct {
	fs.File
	dir   string
	write fs.Filesystem
}

func (tf teefile) Readdir(n int) (fi []os.FileInfo, err error) {
	if fi, err = tf.File.Readdir(n); err != nil {
		return
	}
	for _, fi := range fi {
		path := filepath.Join(tf.dir, filepath.Base(fi.Name()))
		if fi.IsDir() {
			err = tf.write.Mkdir(path, fi.Mode())
		} else {
			var f fs.File
			if f, err = tf.write.Create(path); err == nil {
				f.Close()
			}
		}
	}
	return
}

// TeeFilesystem returns a filesystem which writes file tree read from 'read'
// filesystem in the 'write' one. Every path passed to the Open, Create, Mkdir,
// MkdirAll and Readdir methods is created on the 'write' filesystem only if
// the call was sucessful.
// TeeFilesystem can be used as a spy for recording file and/or directory access
// of the 'read' filesystem.
func TeeFilesystem(read, write fs.Filesystem) fs.Filesystem {
	return teefs{
		read:  read,
		write: write,
	}
}
