package password

import (
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
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			got, err := GetStorePath()

			SetDefaultManager(Managers["old"])

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStorePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetStorePath() got = %v, want %v", got, tt.want)
			}
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
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			err := SetStorePath(tt.args.path)

			SetDefaultManager(Managers["old"])

			if (err != nil) != tt.wantErr {
				t.Errorf("SetStorePath() error = %v, wantErr %v", err, tt.wantErr)
			}
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
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			got, err := GetFileEnding()

			SetDefaultManager(Managers["old"])

			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileEnding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFileEnding() got = %v, want %v", got, tt.want)
			}
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
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			err := SetFileEnding(tt.args.e)

			SetDefaultManager(Managers["old"])

			if (err != nil) != tt.wantErr {
				t.Errorf("SetFileEnding() error = %v, wantErr %v", err, tt.wantErr)
			}
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
			RegisterDefaultManager("old")
			currentManager := GetDefaultManager()
			currentManager.storageBackend = tt.backend

			got, err := FilePath(tt.args.id)

			SetDefaultManager(Managers["old"])

			if (err != nil) != tt.wantErr {
				t.Errorf("FilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FilePath() got = %v, want %v", got, tt.want)
			}
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
			// test init
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

			// test cleanup
			RegisterDefaultManager("old")
		})
	}
}
