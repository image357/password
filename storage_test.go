package password

import (
	"os"
	"path/filepath"
	"reflect"
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
			SetStorePath(tt.args.path)
			if got := GetStorePath(); got != tt.want {
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
			SetFileEnding(tt.args.e)
			if got := GetFileEnding(); got != tt.want {
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
	SetStorePath("mIxEd/Path")
	SetFileEnding("pwd")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilePath(tt.args.id); got != tt.want {
				t.Errorf("FilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestList(t *testing.T) {
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
	SetStorePath("tests/workdir/List")

	// tests
	for _, tt := range tests {
		for _, id := range tt.args.ids {
			err := Overwrite(id, "123", "456")
			if err != nil {
				t.Fatal(err)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := List()
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() got = %v, want %v", got, tt.want)
			}
		})
		err := Clean()
		if err != nil {
			t.Fatal(err)
		}
	}

	// cleanup
	err := os.RemoveAll(GetStorePath())
	if err != nil {
		t.Fatal(err)
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{"a"}, false},
		{"forward slash", args{"b/foo"}, false},
		{"backward slash", args{"c/bar"}, false},
		{"mixed slash", args{"d/foo\\bar/filename"}, false},
		{"invalid id", args{"foobar"}, true},
	}
	// init
	SetStorePath("tests/workdir/Delete")

	err := Overwrite("a", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("b/foo", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("c/bar", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("d/foo/bar/filename", "123", "456")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Delete(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// cleanup
	err = os.RemoveAll(GetStorePath())
	if err != nil {
		t.Fatal(err)
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"normal", false},
		{"empty", false},
	}
	// init
	SetStorePath("tests/workdir/Clean")

	err := Overwrite("a", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("b/foo", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("c/bar", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("d/foo/bar/filename", "123", "456")
	if err != nil {
		t.Fatal(err)
	}

	list, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 4 {
		t.Fatalf("list = %v, want %v", list, []string{"a", "b/foo", "c/bar", "d/foo/bar/filename"})
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Clean(); (err != nil) != tt.wantErr {
				t.Errorf("Clean() error = %v, wantErr %v", err, tt.wantErr)
			}
			list, err = List()
			if err != nil {
				t.Error(err)
			}
			if len(list) != 0 {
				t.Errorf("Clean() list = %v", list)
			}
		})
	}

	// cleanup
	err = os.RemoveAll(GetStorePath())
	if err != nil {
		t.Fatal(err)
	}
}

func Test_packData(t *testing.T) {
	type args struct {
		id   string
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"normal", args{"foo", "bar"}, `{"data":"bar","id":"foo"}`, false},
		{"no escape", args{"foo<>&", "bar<>&"}, `{"data":"bar<>&","id":"foo<>&"}`, false},
		{"empty", args{"", ""}, `{"data":"","id":""}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := packData(tt.args.id, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("packData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("packData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unpackData(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{"normal", args{`{"data": "bar", "id": "foo"}`}, "foo", "bar", false},
		{"no escape", args{`{"data": "bar<>&", "id": "foo<>&"}`}, "foo<>&", "bar<>&", false},
		{"empty", args{`{"data": "", "id": ""}`}, "", "", false},
		{"error", args{``}, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := unpackData(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unpackData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("unpackData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("unpackData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
