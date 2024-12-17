package password

import (
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
		{"enable recovery", args{"123456"}},
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
		{"disable recovery"},
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
