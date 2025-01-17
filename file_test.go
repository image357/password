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
			storePath:            DefaultStorePath,
			storageTree:          map[string]*sync.Mutex{},
			storageTreeLockCount: map[string]int{},
			storageTreeMutex:     sync.Mutex{},
		},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFileStorage()
			want := tt.want
			want.SetStorePath(want.storePath)

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

func TestFileStorage_FilePath(t *testing.T) {
	type fields struct {
		storePath string
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
		{"normal", &fields{storePath: "mIxEd/Path"}, args{"Filename"}, filepath.FromSlash("mIxEd/Path/filename." + DefaultFileEnding)},
		{"forward slash", &fields{storePath: "mIxEd/Path"}, args{"Path/tO/fIle/filEname"}, filepath.FromSlash("mIxEd/Path/path/to/file/filename." + DefaultFileEnding)},
		{"backward slash", &fields{storePath: "mIxEd/Path"}, args{"Path\\tO\\fIle\\filEname"}, filepath.FromSlash("mIxEd/Path/path/to/file/filename." + DefaultFileEnding)},
		{"mixed slash", &fields{storePath: "mIxEd/Path"}, args{"Path/tO\\fIle/filEname"}, filepath.FromSlash("mIxEd/Path/path/to/file/filename." + DefaultFileEnding)},
		{"relative path", &fields{storePath: "mIxEd/Path"}, args{"Path/../tO/fIle/filEname"}, filepath.FromSlash("mIxEd/Path/to/file/filename." + DefaultFileEnding)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileStorage{
				storePath: tt.fields.storePath,
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

			if numLocks := f.storageTreeLockCount[tt.args.id]; numLocks != 1 {
				t.Fatalf("storageTreeLockCount[id] = %v, want %v", numLocks, 1)
			}

			success = f.storageTree[tt.args.id].TryLock()
			if success {
				t.Fatalf("storageTree[id].TryLock() = %v", success)
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

func TestFileStorage_List(t *testing.T) {
	type args struct {
		ids []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{"single", args{[]string{"filename"}}, []string{"filename"}, false},
		{"multi", args{[]string{"a", "c", "b"}}, []string{"a", "b", "c"}, false},
		{"forward slash", args{[]string{"a/foo", "c/bar", "b/baz"}}, []string{"a/foo", "b/baz", "c/bar"}, false},
		{"backward slash", args{[]string{"a\\foo", "c\\bar", "b\\baz"}}, []string{"a/foo", "b/baz", "c/bar"}, false},
		{"mixed slash", args{[]string{"a", "c/bar", "b\\baz"}}, []string{"a", "b/baz", "c/bar"}, false},
	}
	// init
	f := NewFileStorage()
	f.SetStorePath("tests/workdir/FileStorage_List")

	// tests
	for _, tt := range tests {
		// init test
		for _, id := range tt.args.ids {
			err := f.Store(id, "123")
			if err != nil {
				t.Fatal(err)
			}
		}
		// test
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.List()
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() got = %v, want %v", got, tt.want)
			}
		})
		// cleanup test
		err := f.Clean()
		if err != nil {
			t.Fatal(err)
		}
		list, err := f.List()
		if (err != nil) != tt.wantErr {
			t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if len(list) != 0 {
			t.Fatalf("List() got = %v, want empty slice", list)
		}
	}

	// cleanup
	path := f.GetStorePath()
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileStorage_Delete(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"some id", args{"some/id"}, false},
		{"missing id", args{"missing/id"}, true},
	}
	// init
	f := NewFileStorage()
	f.SetStorePath("tests/workdir/FileStorage_List")

	err := f.Store("some/id", "some data")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := f.Delete(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
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

func TestFileStorage_Clean(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"normal", false},
		{"empty", false},
	}
	// init
	f := NewFileStorage()
	f.SetStorePath("tests/workdir/FileStorage_List")

	err := f.Store("some/id", "some data")
	if err != nil {
		t.Fatal(err)
	}

	err = f.Store("another_id", "another data")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := f.Clean(); (err != nil) != tt.wantErr {
				t.Errorf("Clean() error = %v, wantErr %v", err, tt.wantErr)
			}
			list, err := f.List()
			if err != nil {
				t.Fatal(err)
			}
			if len(list) != 0 {
				t.Errorf("Clean() list = %v, want empty", list)
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

func TestFileStorage_DumpJSON(t *testing.T) {
	type fields struct {
		storePath string
		registry  map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"execute", fields{
			storePath: "tests/workdir/FileStorage_DumpJSON",
			registry: map[string]string{
				"a":   "a_data",
				"b/c": "bc_data",
			}},
			`{"a":"a_data","b/c":"bc_data"}`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			f := NewFileStorage()
			f.SetStorePath(tt.fields.storePath)

			for k, v := range tt.fields.registry {
				err := f.Store(k, v)
				if err != nil {
					t.Fatal(err)
				}
			}

			// test
			got, err := f.DumpJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("DumpJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DumpJSON() got = %v, want %v", got, tt.want)
			}

			// cleanup test
			path := f.GetStorePath()
			err = os.RemoveAll(path)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFileStorage_LoadJSON(t *testing.T) {
	type fields struct {
		storePath string
		registry  map[string]string
	}
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantMap map[string]string
		wantErr bool
	}{
		{"success", fields{
			storePath: "tests/workdir/FileStorage_LoadJSON",
			registry: map[string]string{
				"a":   "old_data",
				"b/c": "old_data",
			}},
			args{`{"a":"a_data","b/c":"bc_data","d":"d_data"}`},
			map[string]string{"a": "a_data", "b/c": "bc_data", "d": "d_data"},
			false,
		},

		{"wrong type", fields{
			storePath: "tests/workdir/FileStorage_LoadJSON",
			registry: map[string]string{
				"a":   "old_data",
				"b/c": "old_data",
			}},
			args{`{"a":"a_data","b/c":"bc_data","d":123}`},
			map[string]string{"a": "old_data", "b/c": "old_data"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			f := NewFileStorage()
			f.SetStorePath(tt.fields.storePath)

			for k, v := range tt.fields.registry {
				err := f.Store(k, v)
				if err != nil {
					t.Fatal(err)
				}
			}

			// test
			if err := f.LoadJSON(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("LoadJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			var got = make(map[string]string)
			for id := range tt.wantMap {
				data, err := f.Retrieve(id)
				if err != nil {
					t.Error(err)
				}
				got[id] = data
			}
			if !reflect.DeepEqual(got, tt.wantMap) {
				t.Errorf("storage contents = %v, want %v", got, tt.wantMap)
			}

			// cleanup test
			path := f.GetStorePath()
			err := os.RemoveAll(path)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
