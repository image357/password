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

func TestGet(t *testing.T) {
	type args struct {
		id  string
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"from Overwrite", args{"Foo", "456"}, "123", false},
		{"from Set create", args{"Bar", "def"}, "abc", false},
		{"from Set change", args{"fooBar/Baz", "a2c"}, "foobar", false},
		{"invalid id", args{"fooBar", "a2c"}, "", true},
		{"from Overwrite recovery", args{"Foo" + RecoveryIdSuffix, "recovery_key"}, "456", false},
		{"invalid recovery id", args{"Bar" + RecoveryIdSuffix, "recovery_key"}, "", true},
	}
	// init
	err := SetStorePath("./tests/workdir/Get")
	if err != nil {
		t.Fatal(err)
	}
	oldHashPassword := GetDefaultManager().HashPassword
	GetDefaultManager().HashPassword = false

	EnableRecovery("recovery_key")
	err = Overwrite("foo", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	DisableRecovery()
	err = Set("bar", "", "abc", "def")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("foobar/baz", "123", "a2c")
	if err != nil {
		t.Fatal(err)
	}
	err = Set("foobar/baz", "123", "foobar", "a2c")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.id, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}

	// cleanup
	GetDefaultManager().HashPassword = oldHashPassword
	path, err := GetStorePath()
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheck(t *testing.T) {
	type args struct {
		id       string
		password string
		key      string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
		hashed  bool
	}{
		// no hash
		{"from Overwrite", args{"foo", "123", "456"}, true, false, false},
		{"from Set create", args{"bar", "abc", "def"}, true, false, false},
		{"from Set change true", args{"foobar/baz", "foobar", "a2c"}, true, false, false},
		{"from Set change false", args{"foobar/baz", "wrong", "a2c"}, false, false, false},
		{"invalid id", args{"foobar", "wrong", "a2c"}, false, true, false},
		{"from Overwrite recovery", args{"foo" + RecoveryIdSuffix, "456", "recovery_key"}, true, false, false},

		// hashed passwords
		{"from Overwrite with hash", args{"foo_hash", "123", "456"}, true, false, true},
		{"from Set create with hash", args{"bar_hash", "abc", "def"}, true, false, true},
		{"from Set change true with hash", args{"foobar/baz_hash", "foobar", "a2c"}, true, false, true},
		{"from Set change false with hash", args{"foobar/baz_hash", "wrong", "a2c"}, false, false, true},
		{"invalid id with hash", args{"foobar_hash", "wrong", "a2c"}, false, true, true},
		{"from Overwrite with hash recovery", args{"foo_hash" + RecoveryIdSuffix, "456", "recovery_key"}, true, false, true},
	}
	// init
	err := SetStorePath("./tests/workdir/Check")
	if err != nil {
		t.Fatal(err)
	}
	oldHashPassword := GetDefaultManager().HashPassword
	EnableRecovery("recovery_key")

	GetDefaultManager().HashPassword = false
	err = Overwrite("foo", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Set("bar", "", "abc", "def")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("foobar/baz", "123", "a2c")
	if err != nil {
		t.Fatal(err)
	}
	err = Set("foobar/baz", "123", "foobar", "a2c")
	if err != nil {
		t.Fatal(err)
	}

	GetDefaultManager().HashPassword = true
	err = Overwrite("foo_hash", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Set("bar_hash", "", "abc", "def")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("foobar/baz_hash", "123", "a2c")
	if err != nil {
		t.Fatal(err)
	}
	err = Set("foobar/baz_hash", "123", "foobar", "a2c")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetDefaultManager().HashPassword = tt.hashed
			got, err := Check(tt.args.id, tt.args.password, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Check() got = %v, want %v", got, tt.want)
			}
		})
	}

	// cleanup
	DisableRecovery()
	GetDefaultManager().HashPassword = oldHashPassword
	path, err := GetStorePath()
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSet(t *testing.T) {
	type args struct {
		id          string
		oldPassword string
		newPassword string
		key         string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"create", args{"foo", "", "123", "456"}, false},
		{"change", args{"foo", "123", "789", "456"}, false},
		{"create folder", args{"foo/bar", "", "123", "456"}, false},
		{"change folder", args{"foo/bar", "123", "789", "456"}, false},
		{"create subfolder", args{"bar/baz/foo", "", "456", "abc"}, false},
		{"change subfolder", args{"bar/baz/foo", "456", "789", "abc"}, false},
		{"add subfolder", args{"bar/boo/foo", "", "789", "abc"}, false},
		{"create mixed slashes", args{"forward/backward\\foo", "", "123", "456"}, false},
		{"change mixed slashes", args{"forward\\backward/foo", "123", "789", "456"}, false},
		{"invalid password", args{"foo", "780", "789", "456"}, true},
		{"invalid key", args{"foo", "789", "780", "def"}, true},
		{"valid", args{"foo", "789", "abc", "456"}, false},
	}
	// init
	err := SetStorePath("./tests/workdir/Set")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Set(tt.args.id, tt.args.oldPassword, tt.args.newPassword, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
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

func TestUnset(t *testing.T) {
	type args struct {
		id       string
		password string
		key      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"from Overwrite", args{"foo", "123", "456"}, false},
		{"from Set create", args{"bar", "abc", "def"}, false},
		{"subfolder", args{"foobar/baz", "foobar", "a2c"}, false},
		{"mixed slashes 1", args{"forward/backward\\foo1", "123", "456"}, false},
		{"mixed slashes 2", args{"forward\\backward/foo2", "123", "456"}, false},
		{"invalid id", args{"foobar", "wrong", "a2c"}, true},
		{"invalid password", args{"invalid1", "abc", "456"}, true},
		{"invalid key", args{"invalid2", "123", "abc"}, true},
	}
	// init
	err := SetStorePath("./tests/workdir/Unset")
	if err != nil {
		t.Fatal(err)
	}

	err = Overwrite("foo", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Set("bar", "", "abc", "def")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("foobar/baz", "foobar", "a2c")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("forward/backward/foo1", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("forward/backward/foo2", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("invalid1", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = Overwrite("invalid2", "123", "456")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unset(tt.args.id, tt.args.password, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Unset() error = %v, wantErr %v", err, tt.wantErr)
			}
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

func TestExists(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"success exists", args{"foo"}, true, false},
		{"success not exists", args{"bar"}, false, false},
	}
	// init
	err := SetStorePath("tests/workdir/Exists")
	if err != nil {
		t.Fatal(err)
	}

	err = Overwrite("foo", "123", "456")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Exists(tt.args.id)
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
	path, err := GetStorePath()
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
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
	err := SetStorePath("tests/workdir/List")
	if err != nil {
		t.Fatal(err)
	}

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
	path, err := GetStorePath()
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(path)
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
	err := SetStorePath("tests/workdir/Delete")
	if err != nil {
		t.Fatal(err)
	}

	err = Overwrite("a", "123", "456")
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
	path, err := GetStorePath()
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(path)
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
	err := SetStorePath("tests/workdir/Clean")
	if err != nil {
		t.Fatal(err)
	}

	err = Overwrite("a", "123", "456")
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
	path, err := GetStorePath()
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}
