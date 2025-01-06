package password

import (
	"os"
	"reflect"
	"testing"
)

func Test_SetDefaultManger_GetDefaultManager(t *testing.T) {
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

func Test_EnableHashing_DisableHashing(t *testing.T) {
	type args struct {
		id       string
		password string
		hashing  bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test hash", args{"hashed", "123", true}, true},
		{"empty", args{"empty", "", true}, true},
		{"no hash", args{"hashed", "123", false}, true},
		{"no hash empty", args{"empty", "", false}, true},
	}
	// init
	err := SetStorePath("./tests/workdir/EnableHashing")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			if tt.args.hashing {
				EnableHashing()
			}
			err = Overwrite(tt.args.id, tt.args.password, "456")
			if err != nil {
				t.Errorf("Overwrite() error = %v", err)
			}
			storedPassword, err := Get(tt.args.id, "456")
			if err != nil {
				t.Errorf("Get() error = %v", err)
			}

			// test
			var result = false
			if tt.args.hashing {
				result, err = compareHashedPassword(storedPassword, tt.args.password)
				if err != nil {
					t.Errorf("compareHashedPassword() error = %v", err)
				}
			} else {
				result = comparePassword(storedPassword, tt.args.password)
			}

			if result != tt.want {
				t.Errorf("ToggleHashPassword(): hashes don't match")
			}

			// cleanup test
			DisableHashing()
		})
	}

	// cleanup
	path, err := GetStorePath()
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_EnableRecovery_DisableRecovery(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
	}{
		{"success", args{"123456"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			EnableRecovery(tt.args.key)
			m := GetDefaultManager()
			if !m.withRecovery {
				t.Errorf("GetDefaultManager().withRecovery = false, want true")
			}
			DisableRecovery()
			if m.withRecovery {
				t.Errorf("GetDefaultManager().withRecovery = true, want false")
			}
		})
	}
}

func TestPassword_PublicAPI(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"success"},
	}
	// init
	err := SetStorePath("tests/workdir/Password_PublicAPI")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Overwrite("foo", "123", "456")
			if err != nil {
				t.Fatal(err)
			}

			err = Overwrite("bar", "123", "456")
			if err != nil {
				t.Fatal(err)
			}

			err = Overwrite("foobar", "123", "456")
			if err != nil {
				t.Fatal(err)
			}

			_, err = Get("foo", "456")
			if err != nil {
				t.Fatal(err)
			}

			_, err = Check("foo", "123", "456")
			if err != nil {
				t.Fatal(err)
			}

			err = Set("foo", "123", "789", "456")
			if err != nil {
				t.Fatal(err)
			}

			err = Unset("foo", "789", "456")
			if err != nil {
				t.Fatal(err)
			}

			_, err = Exists("bar")
			if err != nil {
				t.Fatal(err)
			}

			_, err = List()
			if err != nil {
				t.Fatal(err)
			}

			err = Delete("bar")
			if err != nil {
				t.Fatal(err)
			}

			err = Clean()
			if err != nil {
				t.Fatal(err)
			}

			list, err := List()
			if err != nil {
				t.Fatal(err)
			}
			if len(list) != 0 {
				t.Fatalf("List() = %v, want empty", list)
			}
		})
	}

	// cleanup
	path, err := GetStorePath()
	if err != nil {
		t.Fatal(err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRewriteKey(t *testing.T) {
	type args struct {
		id     string
		oldKey string
		newKey string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"success", args{"foobar", "123", "456"}, false},
	}
	// init
	err := SetStorePath("tests/workdir/Password_RewriteKey")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init test
			err := Overwrite(tt.args.id, "password", tt.args.oldKey)
			if err != nil {
				t.Fatal(err)
			}

			// test
			if err := RewriteKey(tt.args.id, tt.args.oldKey, tt.args.newKey); (err != nil) != tt.wantErr {
				t.Errorf("RewriteKey() error = %v, wantErr %v", err, tt.wantErr)
			}

			// cleanup test
			err = Clean()
			if err != nil {
				t.Fatal(err)
			}
		})
	}

	// cleanup
	path, err := GetStorePath()
	if err != nil {
		t.Fatal(err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}
