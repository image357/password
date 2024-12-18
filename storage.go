package password

import (
	"errors"
	pathlib "path"
	"strings"
)

var unsupportedError = errors.New("unsupported storage backend")

type Storage interface {
	Store(id string, data string) error
	Retrieve(id string) (string, error)
	Exists(id string) (bool, error)
	List() ([]string, error)
	Delete(id string) error
	Clean() error
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

	return "", unsupportedError
}

// SetStorePath accepts a new storage path with system-unspecific or mixed path separators.
func SetStorePath(path string) error {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		m.storageBackend.(*FileStorage).SetStorePath(path)
		return nil
	}

	return unsupportedError
}

// GetFileEnding returns the current file ending of storage files.
func GetFileEnding() (string, error) {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		return m.storageBackend.(*FileStorage).GetFileEnding(), nil
	}

	return "", unsupportedError
}

// SetFileEnding accepts a new file ending for storage files.
func SetFileEnding(e string) error {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		m.storageBackend.(*FileStorage).SetFileEnding(e)
		return nil
	}

	return unsupportedError
}

// FilePath returns the storage filepath of a given password-id with system-specific path separators.
// It accepts system-unspecific or mixed id separators, i.e. forward- and backward-slashes are treated as the same character.
func FilePath(id string) (string, error) {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *FileStorage:
		return m.storageBackend.(*FileStorage).FilePath(id), nil
	}

	return "", unsupportedError
}

// SetTemporaryStorage overwrites the current storage backend with a memory based one.
func SetTemporaryStorage() {
	GetDefaultManager().storageBackend = NewTemporaryStorage()
}
