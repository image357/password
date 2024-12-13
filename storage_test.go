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
