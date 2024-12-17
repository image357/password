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
