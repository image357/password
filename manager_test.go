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
	m.storageBackend.(*FileStorage).SetStorePath("./tests/workdir/Overwrite")
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
	m.storageBackend.(*FileStorage).SetStorePath("./tests/workdir/Get")

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
