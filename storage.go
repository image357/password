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
func normalizeSeparator(id string) string {
	return strings.ReplaceAll(id, "\\", "/")
}

// NormalizeId transforms path to lower case letters and normalizes the path separator
func NormalizeId(path string) string {
	path = strings.ToLower(path)
	path = normalizeSeparator(path)
	path = pathlib.Join("/", path)
	path = strings.TrimPrefix(path, "/")
	return pathlib.Clean(path)
}

// GetStorePath returns the current storage path with system-specific path separators.
func GetStorePath() (string, error) {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *fileStorage:
		return m.storageBackend.(*fileStorage).GetStorePath(), nil
	}

	return "", unsupportedError
}

// SetStorePath accepts a new storage path with system-unspecific or mixed path separators.
func SetStorePath(path string) error {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *fileStorage:
		m.storageBackend.(*fileStorage).SetStorePath(path)
		return nil
	}

	return unsupportedError
}

// GetFileEnding returns the current file ending of storage files.
func GetFileEnding() (string, error) {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *fileStorage:
		return m.storageBackend.(*fileStorage).GetFileEnding(), nil
	}

	return "", unsupportedError
}

// SetFileEnding accepts a new file ending for storage files.
func SetFileEnding(e string) error {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *fileStorage:
		m.storageBackend.(*fileStorage).SetFileEnding(e)
		return nil
	}

	return unsupportedError
}

// FilePath returns the storage filepath of a given password-id with system-specific path separators.
// It accepts system-unspecific or mixed id separators, i.e. forward- and backward-slashes are treated as the same character.
func FilePath(id string) (string, error) {
	m := GetDefaultManager()

	switch m.storageBackend.(type) {
	case *fileStorage:
		return m.storageBackend.(*fileStorage).FilePath(id), nil
	}

	return "", unsupportedError
}
