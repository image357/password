package password

import (
	"bytes"
	"encoding/json"
	"errors"
	"sort"
	"strings"
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

// DumpJSON serializes the storage backend to a JSON string.
func (t *TemporaryStorage) DumpJSON() (string, error) {
	// prepare encoder
	temp := new(bytes.Buffer)
	enc := json.NewEncoder(temp)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	// lock storage
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// serialize
	err := enc.Encode(t.registry)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(temp.String(), "\n", ""), nil
}

// LoadJSON deserializes a JSON string into the storage backend.
func (t *TemporaryStorage) LoadJSON(input string) error {
	// prepare decoder
	dec := json.NewDecoder(strings.NewReader(input))
	dec.DisallowUnknownFields()

	// deserialize
	temp := make(map[string]interface{})
	err := dec.Decode(&temp)
	if err != nil {
		return err
	}

	// check value types
	for _, v := range temp {
		switch v.(type) {
		case string:
			// pass
		default:
			return invalidStorageTypeErr
		}
	}

	// lock storage
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// insert data
	for k, v := range temp {
		t.registry[k] = v.(string)
	}

	return nil
}

// WriteToDisk saves the temporary storage to files via FileStorage mechanisms.
// Warning: This method does not block operations on the underlying storage backends (read/write/create/delete).
// You should stop operations manually before usage or ignore the reported error.
// Data consistency is guaranteed.
func (t *TemporaryStorage) WriteToDisk(path string) error {
	f := NewFileStorage()
	f.SetStorePath(path)

	list, err := t.List()
	if err != nil {
		return err
	}

	var lastErr error = nil
	for _, id := range list {
		data, err := t.Retrieve(id)
		if err != nil {
			lastErr = err
			continue
		}

		err = f.Store(id, data)
		if err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return nil
}

// ReadFromDisk loads a FileStorage backend from disk into a temporary storage.
// Warning: This method does not block operations on the underlying storage backends (read/write/create/delete).
// You should stop operations manually before usage or ignore the reported error.
// Data consistency is guaranteed.
func (t *TemporaryStorage) ReadFromDisk(path string) error {
	f := NewFileStorage()
	f.SetStorePath(path)

	list, err := f.List()
	if err != nil {
		return err
	}

	var lastErr error = nil
	for _, id := range list {
		data, err := f.Retrieve(id)
		if err != nil {
			lastErr = err
			continue
		}

		err = t.Store(id, data)
		if err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return nil
}
