package password

import (
	"os"
	"reflect"
	"testing"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name string
		want *Manager
	}{
		{"create", &Manager{
			HashPassword:   false,
			withRecovery:   false,
			storageBackend: NewFileStorage(),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_EnableRecovery(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
	}{
		{"enable", args{"123456"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			if len(m.recoveryKeyBytes) != 0 {
				t.Fatalf("NewManager() should not have any recovery key")
			}
			if len(m.recoveryKeySecret) != 0 {
				t.Fatalf("NewManager() should not have any recovery secret")
			}

			m.withRecovery = false
			m.EnableRecovery(tt.args.key)

			if !m.withRecovery {
				t.Fatalf("manager should have recovery enabled")
			}
			if len(m.recoveryKeyBytes) == 0 {
				t.Errorf("wrong recovery key length = %v", len(m.recoveryKeyBytes))
			}
			if len(m.recoveryKeySecret) == 0 {
				t.Errorf("wrong recovery secret length = %v", len(m.recoveryKeySecret))
			}
		})
	}
}

func TestManager_DisableRecovery(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"disable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()

			m.withRecovery = true
			m.DisableRecovery()

			if m.withRecovery {
				t.Fatalf("manager should not have recovery enabled")
			}
			if len(m.recoveryKeyBytes) != 0 {
				t.Errorf("wrong recovery key length = %v", len(m.recoveryKeyBytes))
			}
			if len(m.recoveryKeySecret) != 0 {
				t.Errorf("wrong recovery secret length = %v", len(m.recoveryKeySecret))
			}
		})
	}
}

func TestManager_getRecoveryKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"get", "123456"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			m.EnableRecovery(tt.want)
			if got := m.getRecoveryKey(); got != tt.want {
				t.Errorf("getRecoveryKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_Overwrite(t *testing.T) {
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
		{"stop recurse on recovery", args{"foo" + RecoveryIdSuffix, "123", "456"}, false},
	}
	// init
	m := NewManager()
	m.storageBackend.(*FileStorage).SetStorePath("./tests/workdir/Manager_Overwrite")
	m.EnableRecovery("recovery_key")

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.Overwrite(tt.args.id, tt.args.password, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Overwrite() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// cleanup
	path := m.storageBackend.(*FileStorage).GetStorePath()
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestManager_Get(t *testing.T) {
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
	m := NewManager()
	m.storageBackend.(*FileStorage).SetStorePath("./tests/workdir/Manager_Get")

	m.EnableRecovery("recovery_key")
	err := m.Overwrite("foo", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	m.DisableRecovery()
	err = m.Set("bar", "", "abc", "def")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Overwrite("foobar/baz", "123", "a2c")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Set("foobar/baz", "123", "foobar", "a2c")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.Get(tt.args.id, tt.args.key)
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
	path := m.storageBackend.(*FileStorage).GetStorePath()
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestManager_Check(t *testing.T) {
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
	m := NewManager()
	m.storageBackend.(*FileStorage).SetStorePath("./tests/workdir/Manager_Check")
	m.EnableRecovery("recovery_key")

	m.HashPassword = false
	err := m.Overwrite("foo", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Set("bar", "", "abc", "def")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Overwrite("foobar/baz", "123", "a2c")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Set("foobar/baz", "123", "foobar", "a2c")
	if err != nil {
		t.Fatal(err)
	}

	m.HashPassword = true
	err = m.Overwrite("foo_hash", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Set("bar_hash", "", "abc", "def")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Overwrite("foobar/baz_hash", "123", "a2c")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Set("foobar/baz_hash", "123", "foobar", "a2c")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.HashPassword = tt.hashed
			got, err := m.Check(tt.args.id, tt.args.password, tt.args.key)
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
	path := m.storageBackend.(*FileStorage).GetStorePath()
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestManager_Set(t *testing.T) {
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
	m := NewManager()
	m.storageBackend.(*FileStorage).SetStorePath("./tests/workdir/Manager_Set")

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.Set(tt.args.id, tt.args.oldPassword, tt.args.newPassword, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// cleanup
	path := m.storageBackend.(*FileStorage).GetStorePath()
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestManager_Unset(t *testing.T) {
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
	m := NewManager()
	m.storageBackend.(*FileStorage).SetStorePath("./tests/workdir/Manager_Unset")

	err := m.Overwrite("foo", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Set("bar", "", "abc", "def")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Overwrite("foobar/baz", "foobar", "a2c")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Overwrite("forward/backward/foo1", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Overwrite("forward/backward/foo2", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Overwrite("invalid1", "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Overwrite("invalid2", "123", "456")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.Unset(tt.args.id, tt.args.password, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Unset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// cleanup
	path := m.storageBackend.(*FileStorage).GetStorePath()
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestManager_Exists(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"success", args{"foo"}, true, false},
		{"invalid id", args{"bar"}, false, false},
	}
	// init
	m := NewManager()
	m.storageBackend.(*FileStorage).SetStorePath("./tests/workdir/Manager_Exists")

	err := m.Overwrite("foo", "123", "456")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.Exists(tt.args.id)
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
	path := m.storageBackend.(*FileStorage).GetStorePath()
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestManager_List(t *testing.T) {
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
	m := NewManager()
	m.storageBackend.(*FileStorage).SetStorePath("./tests/workdir/Manager_List")

	// tests
	for _, tt := range tests {
		// test init
		for _, id := range tt.args.ids {
			err := m.Overwrite(id, "123", "456")
			if err != nil {
				t.Fatal(err)
			}
		}
		// test
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.List()
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() got = %v, want %v", got, tt.want)
			}
		})
		// test cleanup
		err := m.Clean()
		if err != nil {
			t.Fatal(err)
		}
		list, err := m.List()
		if (err != nil) != tt.wantErr {
			t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if len(list) != 0 {
			t.Fatalf("List() got = %v, want empty slice", list)
		}
	}

	// cleanup
	path := m.storageBackend.(*FileStorage).GetStorePath()
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}
