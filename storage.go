package password

import (
	"errors"
	pathlib "path"
	"strings"
)

var unsupportedStorageError = errors.New("unsupported storage backend")
var invalidStorageTypeErr = errors.New("invalid storage type")

type Storage interface {
	// Store (create/overwrite) the provided data.
	Store(id string, data string) error

	// Retrieve data from an existing storage entry.
	Retrieve(id string) (string, error)

	// Exists tests if a given id already exists in the storage backend.
	Exists(id string) (bool, error)

	// List all stored password-ids.
	List() ([]string, error)

	// Delete an existing password.
	Delete(id string) error

	// Clean (delete) all stored passwords.
	Clean() error

	// DumpJSON serializes the storage backend to a JSON string.
	DumpJSON() (string, error)

	// LoadJSON deserializes a JSON string into the storage backend.
	LoadJSON(input string) error
}

// normalizeSeparator replaces all backward-slash ("\\") with forward-slash ("/") characters
func normalizeSeparator(s string) string {
	return strings.ReplaceAll(s, "\\", "/")
}

// NormalizeId transforms path to lower case letters and normalizes the path separator
func NormalizeId(id string) string {
	id = strings.ToLower(id)
	id = normalizeSeparator(id)
	id = pathlib.Join("/", id)
	id = strings.TrimPrefix(id, "/")
	return pathlib.Clean(id)
}

// GetStorePath returns the current storage path with system-specific path separators.
func GetStorePath() (string, error) {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		return m.storageBackend.(*FileStorage).GetStorePath(), nil
	}

	return "", unsupportedStorageError
}

// SetStorePath accepts a new storage path with system-unspecific or mixed path separators.
func SetStorePath(path string) error {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		m.storageBackend.(*FileStorage).SetStorePath(path)
		return nil
	}

	return unsupportedStorageError
}

// GetFileEnding returns the current file ending of storage files.
func GetFileEnding() (string, error) {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		return m.storageBackend.(*FileStorage).GetFileEnding(), nil
	}

	return "", unsupportedStorageError
}

// SetFileEnding accepts a new file ending for storage files.
func SetFileEnding(e string) error {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		m.storageBackend.(*FileStorage).SetFileEnding(e)
		return nil
	}

	return unsupportedStorageError
}

// FilePath returns the storage filepath of a given password-id with system-specific path separators.
// It accepts system-unspecific or mixed id separators, i.e. forward- and backward-slashes are treated as the same character.
func FilePath(id string) (string, error) {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		return m.storageBackend.(*FileStorage).FilePath(id), nil
	}

	return "", unsupportedStorageError
}

// SetTemporaryStorage overwrites the current storage backend with a memory based one.
func SetTemporaryStorage() {
	GetDefaultManager().storageBackend = NewTemporaryStorage()
}

// DumpJSON serializes the storage backend to a JSON string.
func DumpJSON() (string, error) {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		return m.storageBackend.(*FileStorage).DumpJSON()
	case *TemporaryStorage:
		return m.storageBackend.(*TemporaryStorage).DumpJSON()
	}

	return "", unsupportedStorageError
}

// LoadJSON deserializes a JSON string into the storage backend.
func LoadJSON(input string) error {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		return m.storageBackend.(*FileStorage).LoadJSON(input)
	case *TemporaryStorage:
		return m.storageBackend.(*TemporaryStorage).LoadJSON(input)
	}

	return unsupportedStorageError
}

// WriteToDisk saves the current storage to files via FileStorage mechanisms.
// Warning: This method does not block operations on the underlying storage backends (read/write/create/delete).
// You should stop operations manually before usage or ignore the reported error.
// Data consistency is guaranteed.
func WriteToDisk(path string) error {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *TemporaryStorage:
		return m.storageBackend.(*TemporaryStorage).WriteToDisk(path)
	}

	return unsupportedStorageError
}

// ReadFromDisk loads a FileStorage backend from disk into the current storage.
// Warning: This method does not block operations on the underlying storage backends (read/write/create/delete).
// You should stop operations manually before usage or ignore the reported error.
// Data consistency is guaranteed.
func ReadFromDisk(path string) error {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *TemporaryStorage:
		return m.storageBackend.(*TemporaryStorage).ReadFromDisk(path)
	}

	return unsupportedStorageError
}
