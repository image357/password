package password

import (
	"fmt"
	"io/fs"
	"os"
	pathlib "path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

var storePath = pathlib.Clean("./password")
var fileEnding = "pwd"

// StorageFileMode controls the file permission set by this package.
var StorageFileMode os.FileMode = 0600

// StorageDirMode controls the directory permission set by this package.
var StorageDirMode os.FileMode = 0700

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
func GetStorePath() string {
	return filepath.FromSlash(storePath)
}

// SetStorePath accepts a new storage path with system-unspecific or mixed path separators.
func SetStorePath(path string) {
	path = normalizeSeparator(path)
	storePath = pathlib.Clean(path)
}

// GetFileEnding returns the current file ending of storage files.
func GetFileEnding() string {
	return fileEnding
}

// SetFileEnding accepts a new file ending for storage files.
func SetFileEnding(e string) {
	fileEnding = strings.ToLower(strings.TrimPrefix(e, "."))
}

// FilePath returns the storage filepath of a given password-id with system-specific path separators.
// It accepts system-unspecific or mixed id separators, i.e. forward- and backward-slashes are treated as the same character.
func FilePath(id string) string {
	id = NormalizeId(id)
	return filepath.FromSlash(pathlib.Join(storePath, id+"."+fileEnding))
}

// store (create/overwrite) the provided data in a file.
// id is converted to the corresponding filepath.
// If necessary, subfolders are created.
func store(id string, data string) error {
	filePath := FilePath(id)
	folderPath, _ := filepath.Split(filePath)
	if folderPath != "" {
		err := os.MkdirAll(folderPath, StorageDirMode)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, StorageFileMode)
	if err != nil {
		return err
	}

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

// retrieve data from an existing file.
// id is converted to the corresponding filepath.
func retrieve(id string) (string, error) {
	textBytes, err := os.ReadFile(FilePath(id))
	if err != nil {
		return "", err
	}

	if !utf8.Valid(textBytes) {
		return "", fmt.Errorf("invalid utf8 character after file reading")
	}
	text := string(textBytes)

	return text, nil
}

// List all stored password-ids.
func List() ([]string, error) {
	list := make([]string, 0, 16)
	err := filepath.WalkDir(GetStorePath(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), "."+fileEnding) {
			return nil
		}

		path = strings.TrimSuffix(path, "."+fileEnding)
		path, err = filepath.Rel(GetStorePath(), path)
		if err != nil {
			return err
		}
		path = NormalizeId(path)
		list = append(list, path)

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(list)
	return list, nil
}

// Delete an existing password.
func Delete(id string) error {
	err := os.Remove(FilePath(id))
	if err != nil {
		return err
	}
	return nil
}

// Clean (delete) all stored passwords.
func Clean() error {
	list, err := List()
	if err != nil {
		return err
	}

	var lastErr error = nil
	for _, l := range list {
		err = Delete(l)
		if err != nil {
			lastErr = err
		}
	}
	return lastErr
}
