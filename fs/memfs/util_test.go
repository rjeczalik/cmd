package memfs

import "testing"

func TestCompare(t *testing.T) {
	cases := [...]struct {
		lhs FS
		rhs FS
		ok  bool
	}{{
		FS{},
		FS{},
		true,
	}, {
		FS{Tree: Directory{}},
		FS{Tree: Directory{}},
		true,
	}, {
		FS{Tree: Directory{"file": File{}, "dir": Directory{}}},
		FS{Tree: Directory{"file": File{}, "dir": Directory{}}},
		true,
	}, {
		FS{Tree: Directory{"file": File{}, "dir": Directory{}}},
		FS{Tree: Directory{"file": File{}, "dir": Directory{"other": File{}}}},
		false,
	}, {
		FS{Tree: Directory{"file": File{}, "dir": Directory{"other": File{}}}},
		FS{Tree: Directory{"file": File{}, "dir": Directory{}}},
		false,
	}, {
		FS{Tree: Directory{"file": File{}, "dir": Directory{"file": File{}}}},
		FS{Tree: Directory{"file": File{}, "dir": Directory{"other": File{}}}},
		false,
	}, {
		FS{Tree: Directory{"file": File{}, "dir": Directory{"file": File{}}}},
		FS{Tree: Directory{"file": File{}, "dir": Directory{"dir": Directory{}}}},
		false,
	}, {
		FS{Tree: Directory{"file": File{}, "dir": Directory{"dir": Directory{}}}},
		FS{Tree: Directory{"file": File{}, "dir": Directory{"file": File{}}}},
		false,
	}, {
		FS{Tree: Directory{"file1": File{}, "file2": File{}, "file3": File{}}},
		FS{Tree: Directory{"file1": File{}, "file2": File{}, "file3": File{}}},
		true,
	}, {
		FS{Tree: Directory{"dir1": Directory{}, "dir2": Directory{}}},
		FS{Tree: Directory{"dir1": Directory{}, "dir2": Directory{}}},
		true,
	}, {
		FS{Tree: Directory{
			"dir1":  Directory{"file11": File{}, "file12": File{}, "dir13": Directory{}},
			"dir2":  Directory{"file21": File{}, "file22": File{}, "dir23": Directory{}},
			"dir3":  Directory{"file31": File{}, "file32": File{}, "dir33": Directory{}},
			"file4": File{},
		}},
		FS{Tree: Directory{
			"dir1":  Directory{"file11": File{}, "file12": File{}, "dir13": Directory{}},
			"dir2":  Directory{"file21": File{}, "file22": File{}, "dir23": Directory{}},
			"dir3":  Directory{"file31": File{}, "file32": File{}, "dir33": Directory{}},
			"file4": File{},
		}},
		true,
	}, {
		FS{Tree: Directory{
			"dir1":  Directory{"file11": File{}, "file12": File{}, "dir13": Directory{}},
			"dir2":  Directory{"file21": File{}, "file22": File{}, "dir23": Directory{}},
			"dir3":  Directory{"file31": File{}, "file32": File{}, "dir33": Directory{}},
			"file4": File{},
		}},
		FS{Tree: Directory{
			"dir1":  Directory{"file11": File{}, "file12": File{}, "dir13": Directory{}},
			"dir2":  Directory{"file21": File{}, "file22": File{}, "dir23": Directory{}},
			"dir3":  Directory{"file31": File{}, "file32": File{}, "dir33": Directory{"x": File{}}},
			"file4": File{},
		}},
		false,
	}}
	for i, cas := range cases {
		if Compare(cas.lhs, cas.rhs) != cas.ok {
			t.Errorf("want Compare(...)=%v; got %v (i=%d)", cas.ok, !cas.ok, i)
		}
	}
}

func TestFsck(t *testing.T) {
	cases := [...]struct {
		fs FS
		ok bool
	}{{
		FS{},
		true,
	}, {
		FS{Tree: Directory{}},
		true,
	}, {
		FS{Tree: Directory{"x": 1, "file": File{}}},
		false,
	}, {
		FS{Tree: Directory{"dir": Directory{"": File{}}}},
		false,
	}, {
		FS{Tree: Directory{"file1": File{}, "file2": File{}, "x": nil}},
		false,
	}, {
		FS{Tree: Directory{"file1": File{}, "file2": File{}, "x": []byte{}}},
		false,
	}, {
		FS{Tree: Directory{"x": map[string]interface{}{}}},
		false,
	}, {
		FS{Tree: Directory{"dir": Directory{"dir": Directory{"": File{}}}}},
		false,
	}, {
		FS{Tree: Directory{
			"dir1":  Directory{"file11": File{}, "file12": File{}, "dir13": Directory{}},
			"dir2":  Directory{"file21": File{}, "file22": File{}, "dir23": Directory{}},
			"dir3":  Directory{"file31": File{}, "file32": File{}, "dir33": Directory{"file": File{}}},
			"file4": File{},
		}},
		true,
	}}
	for i, cas := range cases {
		if Fsck(cas.fs) != cas.ok {
			t.Errorf("want Fsck(...)=%v; got %v (i=%d)", cas.ok, !cas.ok, i)
		}
	}
}
