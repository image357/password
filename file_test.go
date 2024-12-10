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
				t.Errorf("storePath = %v, want %v", got, tt.want)
			}
		})
	}
}
