package password

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_normalizeSeparator(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"first", args{"first.pwd"}, "first.pwd"},
		{"second", args{"/"}, "/"},
		{"third", args{"//"}, "//"},
		{"fourth", args{"path/fourth.pwd"}, "path/fourth.pwd"},
		{"fifth", args{"path\\fifth.pwd"}, "path/fifth.pwd"},
		{"sixth", args{"\\"}, "/"},
		{"seventh", args{"\\\\"}, "//"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSeparator(tt.args.id); got != tt.want {
				t.Errorf("normalizeSeparator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NormalizeId(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"first", args{"First.pwd"}, "first.pwd"},
		{"second", args{"/"}, "."},
		{"third", args{"//"}, "."},
		{"fourth", args{"pAth/foUrth.pwD"}, "path/fourth.pwd"},
		{"fifth", args{"./Path\\tO/../fiftH.pWd"}, "path/fifth.pwd"},
		{"sixth", args{"\\"}, "."},
		{"seventh", args{"\\\\"}, "."},
		{"eighth", args{"./../.."}, "."},
		{"ninth", args{"./foo/../../to/../../path"}, "path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeId(tt.args.path); got != tt.want {
				t.Errorf("NormalizeId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStorePath(t *testing.T) {
	tests := []struct {
		name    string
		backend Storage
		want    string
		wantErr bool
	}{
		{"pass", &FileStorage{storePath: "foo/bar", fileEnding: "pwd"}, filepath.FromSlash("foo/bar"), false},
		{"fail", nil, filepath.FromSlash(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			// test
			got, err := GetStorePath()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStorePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetStorePath() got = %v, want %v", got, tt.want)
			}

			// cleanup test
			SetDefaultManager(Managers["old"])
		})
	}
}

func TestSetStorePath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		backend Storage
		args    args
		wantErr bool
	}{
		{"pass", NewFileStorage(), args{"some/path"}, false},
		{"fail", nil, args{"some/path"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			// test
			err := SetStorePath(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetStorePath() error = %v, wantErr %v", err, tt.wantErr)
			}

			// cleanup test
			SetDefaultManager(Managers["old"])
		})
	}
}

func TestGetFileEnding(t *testing.T) {
	tests := []struct {
		name    string
		backend Storage
		want    string
		wantErr bool
	}{
		{"pass", &FileStorage{fileEnding: "foobar"}, "foobar", false},
		{"fail", nil, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			// test
			got, err := GetFileEnding()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileEnding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFileEnding() got = %v, want %v", got, tt.want)
			}

			// cleanup test
			SetDefaultManager(Managers["old"])
		})
	}
}

func TestSetFileEnding(t *testing.T) {
	type args struct {
		e string
	}
	tests := []struct {
		name    string
		backend Storage
		args    args
		wantErr bool
	}{
		{"pass", NewFileStorage(), args{"ending"}, false},
		{"fail", nil, args{"ending"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			// test
			err := SetFileEnding(tt.args.e)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetFileEnding() error = %v, wantErr %v", err, tt.wantErr)
			}

			// cleanup test
			SetDefaultManager(Managers["old"])
		})
	}
}

func TestFilePath(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		backend Storage
		args    args
		want    string
		wantErr bool
	}{
		{"pass", &FileStorage{storePath: "some/path", fileEnding: "ending"}, args{"some/id"}, filepath.FromSlash("some/path/some/id.ending"), false},
		{"fail", nil, args{"some/id"}, filepath.FromSlash(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			// test
			got, err := FilePath(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FilePath() got = %v, want %v", got, tt.want)
			}

			// cleanup test
			SetDefaultManager(Managers["old"])
		})
	}
}

func TestSetTemporaryStorage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"execute"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			RegisterDefaultManager("old")

			// test
			_, ok := GetDefaultManager().storageBackend.(*FileStorage)
			if !ok {
				t.Errorf("GetDefaultManager().storageBackend = %v, want FileStorage", GetDefaultManager().storageBackend)
			}

			SetTemporaryStorage()

			_, ok = GetDefaultManager().storageBackend.(*TemporaryStorage)
			if !ok {
				t.Errorf("GetDefaultManager().storageBackend = %v, want TemporaryStorage", GetDefaultManager().storageBackend)
			}

			// cleanup test
			RegisterDefaultManager("old")
		})
	}
}

func TestDumpJSON(t *testing.T) {
	tests := []struct {
		name    string
		backend Storage
		want    string
		wantErr bool
	}{
		{"FileStorage", NewFileStorage(), "{}", false},
		{"TemporaryStorage", NewTemporaryStorage(), "{}", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			switch tt.backend.(type) {
			case *FileStorage:
				tt.backend.(*FileStorage).SetStorePath("./tests/workdir/Storage_DumpJSON")
				path := tt.backend.(*FileStorage).GetStorePath()
				err := os.MkdirAll(path, storageDirMode)
				if err != nil {
					t.Fatal(err)
				}
			}

			// test
			got, err := DumpJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("DumpJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DumpJSON() got = %v, want %v", got, tt.want)
			}

			// cleanup test
			switch tt.backend.(type) {
			case *FileStorage:
				path := tt.backend.(*FileStorage).GetStorePath()
				err = os.RemoveAll(path)
				if err != nil {
					t.Error(err)
				}
			}
			SetDefaultManager(Managers["old"])
		})
	}
}

func TestLoadJSON(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		backend Storage
		args    args
		wantErr bool
	}{
		{"FileStorage", NewFileStorage(), args{"{}"}, false},
		{"TemporaryStorage", NewTemporaryStorage(), args{"{}"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			switch tt.backend.(type) {
			case *FileStorage:
				tt.backend.(*FileStorage).SetStorePath("./tests/workdir/Storage_LoadJSON")
				path := tt.backend.(*FileStorage).GetStorePath()
				err := os.MkdirAll(path, storageDirMode)
				if err != nil {
					t.Fatal(err)
				}
			}

			// test
			if err := LoadJSON(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("LoadJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			// cleanup test
			switch tt.backend.(type) {
			case *FileStorage:
				path := tt.backend.(*FileStorage).GetStorePath()
				err := os.RemoveAll(path)
				if err != nil {
					t.Error(err)
				}
			}
			SetDefaultManager(Managers["old"])
		})
	}
}

func TestWriteToDisk(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		backend Storage
		args    args
		wantErr bool
	}{
		{"FileStorage", NewFileStorage(), args{"./tests/workdir/Storage_WriteToDisk"}, true},
		{"TemporaryStorage", NewTemporaryStorage(), args{"./tests/workdir/Storage_WriteToDisk"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			// test
			if err := WriteToDisk(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("WriteToDisk() error = %v, wantErr %v", err, tt.wantErr)
			}

			// cleanup test
			err := os.RemoveAll(tt.args.path)
			if err != nil {
				t.Error(err)
			}
			SetDefaultManager(Managers["old"])
		})
	}
}

func TestReadFromDisk(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		backend Storage
		args    args
		wantErr bool
	}{
		{"FileStorage", NewFileStorage(), args{"./tests/workdir/Storage_ReadFromDisk"}, true},
		{"TemporaryStorage", NewTemporaryStorage(), args{"./tests/workdir/Storage_ReadFromDisk"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			err := os.MkdirAll(tt.args.path, storageDirMode)
			if err != nil {
				t.Fatal(err)
			}

			// test
			if err := ReadFromDisk(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("ReadFromDisk() error = %v, wantErr %v", err, tt.wantErr)
			}

			// cleanup test
			err = os.RemoveAll(tt.args.path)
			if err != nil {
				t.Error(err)
			}
			SetDefaultManager(Managers["old"])
		})
	}
}
