package password

import (
	"reflect"
	"sync"
	"testing"
)

func TestNewTemporaryStorage(t *testing.T) {
	tests := []struct {
		name string
		want *TemporaryStorage
	}{
		{"create", &TemporaryStorage{
			registry: make(map[string]string),
			mutex:    sync.Mutex{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTemporaryStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTemporaryStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}
