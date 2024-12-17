package password

import (
	"os"
	"reflect"
	"testing"
)

func TestGetDefaultManager(t *testing.T) {
	tests := []struct {
		name string
		want *Manager
	}{
		{"success", NewManager()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultManager(tt.want)
			if got := GetDefaultManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDefaultManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterDefaultManager(t *testing.T) {
	type args struct {
		identifier string
	}
	tests := []struct {
		name string
		args args
	}{
		{"first", args{"first"}},
		{"second", args{"second"}},
		{"third", args{"third"}},
	}
	// init for overwrite demonstration
	RegisterDefaultManager("first")

	// tests
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := GetDefaultManager()
			RegisterDefaultManager(tt.args.identifier)
			if len(Managers) != i+2 {
				t.Errorf("RegisterDefaultManager() length = %v, want %v", len(Managers), i+2)
			}
			if !reflect.DeepEqual(Managers[tt.args.identifier], m) {
				t.Errorf("RegisterDefaultManager() = %v, want %v", Managers[tt.args.identifier], m)
			}
		})
	}
}

func TestToggleHashPassword(t *testing.T) {
	type args struct {
		id       string
		password string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test hash", args{"hashed", "123"}, true},
		{"empty", args{"empty", ""}, true},
	}
	// init
	err := SetStorePath("./tests/workdir/ToggleHashPassword")
	if err != nil {
		t.Fatal(err)
	}
	for ok := ToggleHashPassword(); !ok; ok = ToggleHashPassword() {
		// stop loop when true
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = Overwrite(tt.args.id, tt.args.password, "456")
			if err != nil {
				t.Errorf("Overwrite() error = %v", err)
			}
			storedPassword, err := Get(tt.args.id, "456")
			if err != nil {
				t.Errorf("Get() error = %v", err)
			}
			result, err := compareHashedPassword(storedPassword, tt.args.password)
			if err != nil {
				t.Errorf("compareHashedPassword() error = %v", err)
			}
			if result != tt.want {
				t.Errorf("ToggleHashPassword(): hashes don't match")
			}
		})
	}

	// cleanup
	for ok := ToggleHashPassword(); ok; ok = ToggleHashPassword() {
		// stop loop when false
	}
	path, err := GetStorePath()
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}
