package password

import (
	"os"
	"testing"
)

func TestOverwrite(t *testing.T) {
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
		{"create", args{"foo", "123", "456"}, false},
		{"overwrite", args{"foo", "789", "abc"}, false},
		{"create folder", args{"foo/bar", "123", "456"}, false},
		{"overwrite folder", args{"foo/bar", "789", "abc"}, false},
		{"create subfolder", args{"bar/baz/foo", "123", "456"}, false},
		{"overwrite subfolder", args{"bar/baz/foo", "789", "abc"}, false},
		{"add subfolder", args{"bar/boo/foo", "789", "abc"}, false},
		{"create mixed slashes", args{"forward/backward\\foo", "123", "456"}, false},
		{"overwrite mixed slashes", args{"forward\\backward/foo", "abc", "def"}, false},
	}
	// init
	SetStorePath("./tests/workdir/Overwrite")

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Overwrite(tt.args.id, tt.args.password, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Overwrite() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// cleanup
	err := os.RemoveAll(GetStorePath())
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
		{"from Overwrite recovery", args{"Foo.recovery", "recovery_key"}, "456", false},
		{"invalid recovery id", args{"Bar.recovery", "recovery_key"}, "", true},
	}
	// init
	SetStorePath("./tests/workdir/Get")
	oldHashPassword := HashPassword
	HashPassword = false

	EnableRecovery("recovery_key")
	err := Overwrite("foo", "123", "456")
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
	HashPassword = oldHashPassword
	err = os.RemoveAll(GetStorePath())
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

		// hashed passwords
		{"from Overwrite with hash", args{"foo_hash", "123", "456"}, true, false, true},
		{"from Set create with hash", args{"bar_hash", "abc", "def"}, true, false, true},
		{"from Set change true with hash", args{"foobar/baz_hash", "foobar", "a2c"}, true, false, true},
		{"from Set change false with hash", args{"foobar/baz_hash", "wrong", "a2c"}, false, false, true},
		{"invalid id with hash", args{"foobar_hash", "wrong", "a2c"}, false, true, true},
	}
	// init
	SetStorePath("./tests/workdir/Check")
	oldHashPassword := HashPassword

	HashPassword = false
	err := Overwrite("foo", "123", "456")
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

	HashPassword = true
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
			HashPassword = tt.hashed
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
	HashPassword = oldHashPassword
	err = os.RemoveAll(GetStorePath())
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
	SetStorePath("./tests/workdir/Set")

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Set(tt.args.id, tt.args.oldPassword, tt.args.newPassword, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// cleanup
	err := os.RemoveAll(GetStorePath())
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
	SetStorePath("./tests/workdir/Unset")

	err := Overwrite("foo", "123", "456")
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
	err = os.RemoveAll(GetStorePath())
	if err != nil {
		t.Fatal(err)
	}
}
