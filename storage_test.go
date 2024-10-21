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

func Test_normalizePath(t *testing.T) {
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
			err := SetStorePath(tt.args.path)
			if err != nil {
				t.Fatal(err)
			}
			expected, _ := filepath.Abs(tt.want)
			got, err := GetStorePath()
			if err != nil {
				t.Fatal(err)
			}
			if got != expected {
				t.Errorf("GetStorePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFileEnding(t *testing.T) {
	type args struct {
		e string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{"pwd"}, "pwd"},
		{"mixed caps", args{"PwD"}, "pwd"},
		{"dot prefix", args{".pwd"}, "pwd"},
		{"double prefix", args{"..pwd"}, ".pwd"},
		{"mixed dot", args{".pwd.old"}, "pwd.old"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetFileEnding(tt.args.e)
			if err != nil {
				t.Fatal(err)
			}
			got, err := GetFileEnding()
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("GetFileEnding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilePath(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{"Filename"}, filepath.FromSlash("mIxEd/Path/filename.pwd")},
		{"forward slash", args{"Path/tO/fIle/filEname"}, filepath.FromSlash("mIxEd/Path/path/to/file/filename.pwd")},
		{"backward slash", args{"Path\\tO\\fIle\\filEname"}, filepath.FromSlash("mIxEd/Path/path/to/file/filename.pwd")},
		{"mixed slash", args{"Path/tO\\fIle/filEname"}, filepath.FromSlash("mIxEd/Path/path/to/file/filename.pwd")},
		{"relative path", args{"Path/../tO/fIle/filEname"}, filepath.FromSlash("mIxEd/Path/to/file/filename.pwd")},
	}
	// init
	err := SetStorePath("mIxEd/Path")
	if err != nil {
		t.Fatal(err)
	}
	err = SetFileEnding("pwd")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected, _ := filepath.Abs(tt.want)
			got, err := FilePath(tt.args.id)
			if err != nil {
				t.Fatal(err)
			}
			if got != expected {
				t.Errorf("FilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
