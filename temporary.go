package password

import (
	"errors"
	"sort"
	"sync"
)

var invalidTemporaryStorageIdErr = errors.New("invalid temporary storage id")

// TemporaryStorage is a memory based storage backend.
type TemporaryStorage struct {
	registry map[string]string
	mutex    sync.Mutex
}

// NewTemporaryStorage returns a memory based storage backend.
func NewTemporaryStorage() *TemporaryStorage {
	t := new(TemporaryStorage)
	t.registry = make(map[string]string)
	return t
}

// Store (create/overwrite) the provided data.
func (t *TemporaryStorage) Store(id string, data string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.registry[id] = data

	return nil
}

// Retrieve data from an existing memory location.
func (t *TemporaryStorage) Retrieve(id string) (string, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	data, ok := t.registry[id]
	if !ok {
		delete(t.registry, id)
		return "", invalidTemporaryStorageIdErr
	}

	return data, nil
}

// Exists tests if a given id already exists in the storage backend.
func (t *TemporaryStorage) Exists(id string) (bool, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	_, ok := t.registry[id]
	if !ok {
		delete(t.registry, id)
		return false, nil
	}

	return true, nil
}

// List all stored password-ids.
func (t *TemporaryStorage) List() ([]string, error) {
	list := make([]string, 0, 16)

	t.mutex.Lock()
	defer t.mutex.Unlock()

	for id := range t.registry {
		list = append(list, id)
	}

	sort.Strings(list)
	return list, nil
}

// Delete an existing password.
func (t *TemporaryStorage) Delete(id string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	_, ok := t.registry[id]
	if !ok {
		delete(t.registry, id)
		return invalidTemporaryStorageIdErr
	}

	delete(t.registry, id)
	return nil
}

// Clean (delete) all stored passwords.
func (t *TemporaryStorage) Clean() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.registry = make(map[string]string)

	return nil
}
