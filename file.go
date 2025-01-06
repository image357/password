package password

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/image357/password/log"
	"io/fs"
	"os"
	pathlib "path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

// DefaultStorePath is the default relative storage path of a file storage backend.
const DefaultStorePath = "./password"

// DefaultFileEnding is the default file extension for password files.
const DefaultFileEnding string = "pwd"

// storageFileMode controls the file permission set by this package.
const storageFileMode os.FileMode = 0600

// storageDirMode controls the directory permission set by this package.
const storageDirMode os.FileMode = 0700

// FileStorage is a file based storage backend.
type FileStorage struct {
	// storePath holds the absolute storage path.
	storePath string

	// storageTree holds an id to sync.Mutex map for thread-safe file access.
	storageTree map[string]*sync.Mutex

	// storageTreeLockCount holds an id to count map for cleaning up unused sync.Mutex entries in storageTree.
	storageTreeLockCount map[string]int

	// storageTreeMutex controls thread-safe access to the storageTree.
	storageTreeMutex sync.Mutex
}

// NewFileStorage returns a default initialized storage backend for persistent files.
func NewFileStorage() *FileStorage {
	f := new(FileStorage)

	f.SetStorePath(DefaultStorePath)
	f.storageTree = make(map[string]*sync.Mutex)
	f.storageTreeLockCount = make(map[string]int)

	return f
}

// GetStorePath returns the current storage path with system-specific path separators.
func (f *FileStorage) GetStorePath() string {
	return filepath.FromSlash(f.storePath)
}

// SetStorePath accepts a new storage path with system-unspecific or mixed path separators.
func (f *FileStorage) SetStorePath(path string) {
	temp, err := filepath.Abs(path)
	if err != nil {
		log.Warn("cannot resolve absolute storage path", "path", path)
	} else {
		path = temp
	}
	path = normalizeSeparator(path)
	f.storePath = pathlib.Clean(path)
}

// FilePath returns the storage filepath of a given password-id with system-specific path separators.
// It accepts system-unspecific or mixed id separators, i.e. forward- and backward-slashes are treated as the same character.
func (f *FileStorage) FilePath(id string) string {
	id = NormalizeId(id)
	return filepath.FromSlash(pathlib.Join(f.storePath, id+"."+DefaultFileEnding))
}

// lockId locks a storage id mutex by first locking the storage tree and increasing lock count.
func (f *FileStorage) lockId(id string) {
	id = NormalizeId(id)

	// get mutex with side effects (create if necessary)
	f.storageTreeMutex.Lock()
	idMutex, ok := f.storageTree[id]
	if !ok {
		idMutex = &sync.Mutex{}
		f.storageTree[id] = idMutex
	}
	f.storageTreeLockCount[id]++
	f.storageTreeMutex.Unlock()

	// lock mutex
	idMutex.Lock()
}

// unlockId locks a storage id mutex by first locking the storage tree and decreasing lock count.
// The storage tree is cleaned from id if lock count is zero.
func (f *FileStorage) unlockId(id string) {
	id = NormalizeId(id)

	// try get mutex without side effects
	f.storageTreeMutex.Lock()
	idMutex, ok := f.storageTree[id]
	if !ok {
		delete(f.storageTree, id)
		delete(f.storageTreeLockCount, id)
	}
	f.storageTreeMutex.Unlock()

	if !ok {
		// abort on missing mutex
		return
	} else {
		// unlock mutex
		idMutex.Unlock()
	}

	// cleanup if last lock
	f.storageTreeMutex.Lock()
	f.storageTreeLockCount[id]--
	if f.storageTreeLockCount[id] <= 0 {
		delete(f.storageTree, id)
		delete(f.storageTreeLockCount, id)
	}
	f.storageTreeMutex.Unlock()
}

// Store (create/overwrite) the provided data in a file.
// id is converted to the corresponding filepath.
// If necessary, subfolders are created.
func (f *FileStorage) Store(id string, data string) error {
	filePath := f.FilePath(id)
	folderPath, _ := filepath.Split(filePath)
	if folderPath != "" {
		err := os.MkdirAll(folderPath, storageDirMode)
		if err != nil {
			return err
		}
	}

	f.lockId(id)
	defer f.unlockId(id)

	err := os.WriteFile(filePath, []byte(data), storageFileMode)
	if err != nil {
		_ = os.Remove(filePath)
		return err
	}

	return nil
}

// Retrieve data from an existing file.
// id is converted to the corresponding filepath.
func (f *FileStorage) Retrieve(id string) (string, error) {
	f.lockId(id)
	defer f.unlockId(id)

	textBytes, err := os.ReadFile(f.FilePath(id))
	if err != nil {
		return "", err
	}

	if !utf8.Valid(textBytes) {
		return "", fmt.Errorf("invalid utf8 character after file reading")
	}
	text := string(textBytes)

	return text, nil
}

// Exists tests if a given id already exists in the storage backend.
func (f *FileStorage) Exists(id string) (bool, error) {
	_, err := os.Stat(f.FilePath(id))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// List all stored password-ids.
func (f *FileStorage) List() ([]string, error) {
	list := make([]string, 0, 16)
	err := filepath.WalkDir(f.GetStorePath(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), "."+DefaultFileEnding) {
			return nil
		}

		path = strings.TrimSuffix(path, "."+DefaultFileEnding)
		path, err = filepath.Rel(f.GetStorePath(), path)
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
func (f *FileStorage) Delete(id string) error {
	f.lockId(id)
	defer f.unlockId(id)

	err := os.Remove(f.FilePath(id))
	if err != nil {
		return err
	}

	return nil
}

// Clean (delete) all stored passwords.
func (f *FileStorage) Clean() error {
	list, err := f.List()
	if err != nil {
		return err
	}

	var lastErr error = nil
	for _, l := range list {
		err = f.Delete(l)
		if err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// DumpJSON serializes the storage backend to a JSON string.
// Warning: This method does not block operations on the underlying storage backend (read/write/create/delete).
// You should stop operations manually before usage or ignore the reported error.
// Data consistency is guaranteed.
func (f *FileStorage) DumpJSON() (string, error) {
	// prepare encoder
	temp := new(bytes.Buffer)
	enc := json.NewEncoder(temp)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	// get ids
	list, err := f.List()
	if err != nil {
		return "", err
	}

	// loop storage
	var lastErr error = nil
	var registry = make(map[string]string)
	for _, id := range list {
		data, err := f.Retrieve(id)
		if err != nil {
			lastErr = err
			continue
		}
		registry[id] = data
	}

	// serialize
	err = enc.Encode(registry)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(temp.String(), "\n", ""), lastErr
}

// LoadJSON deserializes a JSON string into the storage backend.
// Warning: This method does not block operations on the underlying storage backend (read/write/create/delete).
// You should stop operations manually before usage or ignore the reported error.
// Data consistency is guaranteed.
func (f *FileStorage) LoadJSON(input string) error {
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

	// write data files
	var lastErr error = nil
	for k, v := range temp {
		err := f.Store(k, v.(string))
		if err != nil {
			lastErr = err
		}
	}

	return lastErr
}
