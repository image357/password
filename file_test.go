package password

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestNewFileStorage(t *testing.T) {
	tests := []struct {
		name string
		want *FileStorage
	}{
		{"create", &FileStorage{
			DefaultStorePath,
			DefaultFileEnding,
			map[string]*sync.Mutex{},
			map[string]int{},
			sync.Mutex{},
		},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFileStorage()
			want := tt.want
			want.SetStorePath(want.storePath)
			want.SetFileEnding(want.fileEnding)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("NewFileStorage() = %v, want %v", got, want)
			}
		})
	}
}

func TestFileStorage_GetStorePath(t *testing.T) {
	type fields struct {
		storePath string
	}
	tests := []struct {
		name   string
		fields *fields
		want   string
	}{
		{"forward slash", &fields{storePath: "/"}, string(os.PathSeparator)},
		{"backward slash", &fields{storePath: "\\"}, "\\"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileStorage{
				storePath: tt.fields.storePath,
			}
			if got := f.GetStorePath(); got != tt.want {
				t.Errorf("GetStorePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStorage_SetStorePath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"current directory 1", args{"."}, "."},
		{"current directory 2", args{"./"}, "."},
		{"current directory 3", args{".\\"}, "."},
		{"previous directory 1", args{".."}, ".."},
		{"previous directory 2", args{"../"}, ".."},
		{"previous directory 3", args{"..\\"}, ".."},
		{"previous directory 4", args{"./.."}, ".."},
		{"previous directory 5", args{"../.."}, filepath.FromSlash("../..")},
		{"some directory 1", args{"Path/tO/sTore"}, filepath.FromSlash("Path/tO/sTore")},
		{"some directory 2", args{"Path/tO\\sTore"}, filepath.FromSlash("Path/tO/sTore")},
		{"some directory 3", args{"Path\\tO/sTore"}, filepath.FromSlash("Path/tO/sTore")},
		{"absolute path 1", args{"/path/to/store"}, filepath.FromSlash("/path/to/store")},
		{"absolute path 2", args{"C:\\path\\to\\store"}, filepath.FromSlash("C:/path/to/store")},
		{"some directory 4", args{"./path/to/store"}, filepath.FromSlash("path/to/store")},
		{"mixed directory 1", args{"../path/to/store"}, filepath.FromSlash("../path/to/store")},
		{"mixed directory 2", args{"./../path/to/store"}, filepath.FromSlash("../path/to/store")},
		{"mixed directory 3", args{"path/../to/store"}, filepath.FromSlash("to/store")},
		{"mixed directory 4", args{"/path/../to/store"}, filepath.FromSlash("/to/store")},
		{"mixed directory 5", args{"C:\\path\\..\\to\\store"}, filepath.FromSlash("C:/to/store")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileStorage()
			f.SetStorePath(tt.args.path)

			expected, err := filepath.Abs(tt.want)
			if err != nil {
				t.Fatal(err)
			}
			expected = normalizeSeparator(expected)
			got := f.storePath

			if got != expected {
				t.Errorf("f.storePath = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStorage_GetFileEnding(t *testing.T) {
	type fields struct {
		fileEnding string
	}
	tests := []struct {
		name   string
		fields *fields
		want   string
	}{
		{"lower case", &fields{fileEnding: "foobar"}, "foobar"},
		{"upper case", &fields{fileEnding: "FOOBAR"}, "FOOBAR"},
		{"mixed case", &fields{fileEnding: "FooBar"}, "FooBar"},
		{"prefix", &fields{fileEnding: ".foobar"}, ".foobar"},
		{"double prefix", &fields{fileEnding: "..foobar"}, "..foobar"},
		{"suffix", &fields{fileEnding: "foobar."}, "foobar."},
		{"double suffix", &fields{fileEnding: "foobar.."}, "foobar.."},
		{"prefix and suffix", &fields{fileEnding: ".foobar."}, ".foobar."},
		{"mixed dot", &fields{fileEnding: "foo.bar"}, "foo.bar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileStorage{
				fileEnding: tt.fields.fileEnding,
			}
			if got := f.GetFileEnding(); got != tt.want {
				t.Errorf("GetFileEnding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStorage_SetFileEnding(t *testing.T) {
	type args struct {
		e string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"lower case", args{e: "foobar"}, "foobar"},
		{"upper case", args{e: "FOOBAR"}, "foobar"},
		{"mixed case", args{e: "FooBar"}, "foobar"},
		{"prefix", args{e: ".foobar"}, "foobar"},
		{"double prefix", args{e: "..foobar"}, "foobar"},
		{"suffix", args{e: "foobar."}, "foobar"},
		{"double suffix", args{e: "foobar.."}, "foobar"},
		{"prefix and suffix", args{e: ".foobar."}, "foobar"},
		{"mixed dot", args{e: "foo.bar"}, "foo.bar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileStorage()
			f.SetFileEnding(tt.args.e)

			if got := f.fileEnding; got != tt.want {
				t.Errorf("f.fileEnding = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStorage_FilePath(t *testing.T) {
	type fields struct {
		storePath  string
		fileEnding string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields *fields
		args   args
		want   string
	}{
		{"normal", &fields{storePath: "mIxEd/Path", fileEnding: "pwd"}, args{"Filename"}, filepath.FromSlash("mIxEd/Path/filename.pwd")},
		{"forward slash", &fields{storePath: "mIxEd/Path", fileEnding: "pwd"}, args{"Path/tO/fIle/filEname"}, filepath.FromSlash("mIxEd/Path/path/to/file/filename.pwd")},
		{"backward slash", &fields{storePath: "mIxEd/Path", fileEnding: "pwd"}, args{"Path\\tO\\fIle\\filEname"}, filepath.FromSlash("mIxEd/Path/path/to/file/filename.pwd")},
		{"mixed slash", &fields{storePath: "mIxEd/Path", fileEnding: "pwd"}, args{"Path/tO\\fIle/filEname"}, filepath.FromSlash("mIxEd/Path/path/to/file/filename.pwd")},
		{"relative path", &fields{storePath: "mIxEd/Path", fileEnding: "pwd"}, args{"Path/../tO/fIle/filEname"}, filepath.FromSlash("mIxEd/Path/to/file/filename.pwd")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileStorage{
				storePath:  tt.fields.storePath,
				fileEnding: tt.fields.fileEnding,
			}
			if got := f.FilePath(tt.args.id); got != tt.want {
				t.Errorf("FilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStorage_lockId(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
	}{
		{"pass", args{"some/id"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileStorage()

			// first lock
			f.lockId(tt.args.id)

			// tests
			success := f.storageTreeMutex.TryLock()
			if !success {
				t.Fatalf("storageTreeMutex.TryLock() = %v", success)
			}
			f.storageTreeMutex.Unlock()

			if lenStorageTree := len(f.storageTree); lenStorageTree != 1 {
				t.Fatalf("len(f.storageTree) = %v, want %v", lenStorageTree, 1)
			}

			if lenStorageTreeLockCount := len(f.storageTreeLockCount); lenStorageTreeLockCount != 1 {
				t.Fatalf("len(f.storageTreeLockCount) = %v, want %v", lenStorageTreeLockCount, 1)
			}

			if numLocks := f.storageTreeLockCount[NormalizeId(tt.args.id)]; numLocks != 1 {
				t.Fatalf("storageTreeLockCount[NormalizeId(id)] = %v, want %v", numLocks, 1)
			}

			success = f.storageTree[NormalizeId(tt.args.id)].TryLock()
			if success {
				t.Fatalf("storageTree[NormalizeId(id)].TryLock() = %v", success)
			}
		})
	}
}

func TestFileStorage_unlockId(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
	}{
		{"pass", args{"some/id"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileStorage()

			// initial tests
			success := f.storageTreeMutex.TryLock()
			if !success {
				t.Fatalf("storageTreeMutex.TryLock() = %v", success)
			}
			f.storageTreeMutex.Unlock()

			if lenStorageTree := len(f.storageTree); lenStorageTree != 0 {
				t.Fatalf("len(f.storageTree) = %v, want %v", lenStorageTree, 0)
			}

			if lenStorageTreeLockCount := len(f.storageTreeLockCount); lenStorageTreeLockCount != 0 {
				t.Fatalf("len(f.storageTreeLockCount) = %v, want %v", lenStorageTreeLockCount, 0)
			}

			// unlock without previous lock
			f.unlockId(tt.args.id)

			// first unlock tests
			success = f.storageTreeMutex.TryLock()
			if !success {
				t.Fatalf("storageTreeMutex.TryLock() = %v", success)
			}
			f.storageTreeMutex.Unlock()

			if lenStorageTree := len(f.storageTree); lenStorageTree != 0 {
				t.Fatalf("len(f.storageTree) = %v, want %v", lenStorageTree, 0)
			}

			if lenStorageTreeLockCount := len(f.storageTreeLockCount); lenStorageTreeLockCount != 0 {
				t.Fatalf("len(f.storageTreeLockCount) = %v, want %v", lenStorageTreeLockCount, 0)
			}

			// lock then unlock
			f.lockId(tt.args.id)
			f.unlockId(tt.args.id)

			// second unlock tests
			success = f.storageTreeMutex.TryLock()
			if !success {
				t.Fatalf("storageTreeMutex.TryLock() = %v", success)
			}
			f.storageTreeMutex.Unlock()

			if lenStorageTree := len(f.storageTree); lenStorageTree != 0 {
				t.Fatalf("len(f.storageTree) = %v, want %v", lenStorageTree, 0)
			}

			if lenStorageTreeLockCount := len(f.storageTreeLockCount); lenStorageTreeLockCount != 0 {
				t.Fatalf("len(f.storageTreeLockCount) = %v, want %v", lenStorageTreeLockCount, 0)
			}
		})
	}
}

func TestFileStorage_Store(t *testing.T) {
	type args struct {
		id   string
		data string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"create", args{"some/id", "some data"}, false},
		{"overwrite", args{"some/id", "another data"}, false},
		{"create another", args{"another/id", "another data"}, false},
	}
	// init
	f := NewFileStorage()
	f.SetStorePath("tests/workdir/FileStorage_Store")

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := f.Store(tt.args.id, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}

			bytes, err := os.ReadFile(f.FilePath(tt.args.id))
			if (err != nil) != tt.wantErr {
				t.Errorf("os.ReadFile() error = %v", err)
			}

			if string(bytes) != tt.args.data {
				t.Errorf("os.ReadFile() = %v, want %v", string(bytes), tt.args.data)
			}
		})
	}

	// cleanup
	path := f.GetStorePath()
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileStorage_Retrieve(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"some id", args{"some/id"}, "some data", false},
		{"another id", args{"another/id"}, "another data", false},
		{"missing id", args{"missing/id"}, "", true},
	}
	// init
	f := NewFileStorage()
	f.SetStorePath("tests/workdir/FileStorage_Retrieve")

	err := f.Store("some/id", "some data")
	if err != nil {
		t.Fatal(err)
	}

	err = f.Store("another/id", "another data")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.Retrieve(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Retrieve() got = %v, want %v", got, tt.want)
			}
		})
	}

	// cleanup
	path := f.GetStorePath()
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileStorage_Exists(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"some id", args{"some/id"}, true, false},
		{"another id", args{"another/id"}, true, false},
		{"missing id", args{"missing/id"}, false, false},
	}
	// init
	f := NewFileStorage()
	f.SetStorePath("tests/workdir/FileStorage_Exists")

	err := f.Store("some/id", "some data")
	if err != nil {
		t.Fatal(err)
	}

	err = f.Store("another/id", "some data")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.Exists(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exists() got = %v, want %v", got, tt.want)
			}
		})
	}

	// cleanup
	path := f.GetStorePath()
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}
