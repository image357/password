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

func TestTemporaryStorage_Retrieve(t1 *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"first", args{"first"}, "1", false},
		{"second", args{"second"}, "2", false},
		{"third", args{"third"}, "3", false},
		{"invalid id", args{"invalid"}, "", true},
	}
	// init
	t := NewTemporaryStorage()

	err := t.Store("first", "1")
	if err != nil {
		t1.Fatal(err)
	}

	err = t.Store("second", "2")
	if err != nil {
		t1.Fatal(err)
	}

	err = t.Store("third", "3")
	if err != nil {
		t1.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			got, err := t.Retrieve(tt.args.id)
			if (err != nil) != tt.wantErr {
				t1.Errorf("Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("Retrieve() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemporaryStorage_Exists(t1 *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"first", args{"first"}, true, false},
		{"second", args{"second"}, true, false},
		{"third", args{"third"}, true, false},
		{"invalid id", args{"invalid"}, false, false},
	}
	// init
	t := NewTemporaryStorage()

	err := t.Store("first", "1")
	if err != nil {
		t1.Fatal(err)
	}

	err = t.Store("second", "2")
	if err != nil {
		t1.Fatal(err)
	}

	err = t.Store("third", "3")
	if err != nil {
		t1.Fatal(err)
	}

	// tests
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			got, err := t.Exists(tt.args.id)
			if (err != nil) != tt.wantErr {
				t1.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("Exists() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemporaryStorage_List(t1 *testing.T) {
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
	}
	// init
	t := NewTemporaryStorage()

	// tests
	for _, tt := range tests {
		// init test
		for _, id := range tt.args.ids {
			err := t.Store(id, "123")
			if err != nil {
				t1.Fatal(err)
			}
		}
		// test
		t1.Run(tt.name, func(t1 *testing.T) {
			got, err := t.List()
			if (err != nil) != tt.wantErr {
				t1.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("List() got = %v, want %v", got, tt.want)
			}
		})
		// cleanup test
		err := t.Clean()
		if err != nil {
			t1.Fatal(err)
		}
		list, err := t.List()
		if (err != nil) != tt.wantErr {
			t1.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if len(list) != 0 {
			t1.Fatalf("List() got = %v, want empty slice", list)
		}
	}
}
