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

func TestTemporaryStorage_Store(t1 *testing.T) {
	type args struct {
		id   string
		data string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"first", args{id: "first", data: "first"}, false},
		{"second", args{id: "second", data: "second"}, false},
		{"third", args{id: "third", data: "third"}, false},
	}
	// init
	t := NewTemporaryStorage()

	// tests
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			if err := t.Store(tt.args.id, tt.args.data); (err != nil) != tt.wantErr {
				t1.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
			_, ok := t.registry[tt.args.id]
			if !ok {
				t1.Errorf("id not found: %v", tt.args.id)
			}
		})
	}
}
