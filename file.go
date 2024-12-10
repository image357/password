package password

import (
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

// DefaultFileEnding is the default file ending for password files of a file storage backend.
const DefaultFileEnding string = "pwd"

// storageFileMode controls the file permission set by this package.
const storageFileMode os.FileMode = 0600

// storageDirMode controls the directory permission set by this package.
const storageDirMode os.FileMode = 0700

// FileStorage is a file based storage backend.
type FileStorage struct {
	// storePath holds the absolute storage path.
	storePath string

	// fileEnding holds the file ending without dot prefix.
	fileEnding string

	// storageTree holds an id to sync.Mutex map for thread-safe file access.
	storageTree map[string]*sync.Mutex

	// storageTreeLockCount holds an id to count map for cleaning up unused sync.Mutex entries in storageTree.
	storageTreeLockCount map[string]int

	// storageTreeMutex controls thread-safe access to the storageTree.
	storageTreeMutex sync.Mutex
}

func NewFileStorage() *FileStorage {
	f := new(FileStorage)

	f.SetStorePath(DefaultStorePath)
	f.SetFileEnding(DefaultFileEnding)
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

// GetFileEnding returns the current file ending of storage files.
func (f *FileStorage) GetFileEnding() string {
	return f.fileEnding
}

// SetFileEnding accepts a new file ending for storage files.
func (f *FileStorage) SetFileEnding(e string) {
	f.fileEnding = strings.ToLower(strings.Trim(e, "."))
}

// FilePath returns the storage filepath of a given password-id with system-specific path separators.
// It accepts system-unspecific or mixed id separators, i.e. forward- and backward-slashes are treated as the same character.
func (f *FileStorage) FilePath(id string) string {
	id = NormalizeId(id)
	return filepath.FromSlash(pathlib.Join(f.storePath, id+"."+f.fileEnding))
}

// lockId locks a storage id mutex by first locking the storage tree and increasing lock count.
func (f *FileStorage) lockId(id string) {
	id = NormalizeId(id)
	f.storageTreeMutex.Lock()

	idMutex, ok := f.storageTree[id]
	if !ok {
		idMutex = &sync.Mutex{}
		f.storageTree[id] = idMutex
	}
	idMutex.Lock()
	f.storageTreeLockCount[id]++

	f.storageTreeMutex.Unlock()
}

// unlockId locks a storage id mutex by first locking the storage tree and decreasing lock count.
// The storage tree is cleaned from id if lock count is zero.
func (f *FileStorage) unlockId(id string) {
	id = NormalizeId(id)
	f.storageTreeMutex.Lock()

	idMutex, ok := f.storageTree[id]
	if !ok {
		delete(f.storageTree, id)
		delete(f.storageTreeLockCount, id)
		return
	}

	idMutex.Unlock()
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
	err := os.WriteFile(filePath, []byte(data), storageFileMode)
	if err != nil {
		_ = os.Remove(filePath)
		f.unlockId(id)
		return err
	}
	f.unlockId(id)

	return nil
}

// Retrieve data from an existing file.
// id is converted to the corresponding filepath.
func (f *FileStorage) Retrieve(id string) (string, error) {
	f.lockId(id)
	textBytes, err := os.ReadFile(f.FilePath(id))
	f.unlockId(id)
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

		if !strings.HasSuffix(d.Name(), "."+f.GetFileEnding()) {
			return nil
		}

		path = strings.TrimSuffix(path, "."+f.GetFileEnding())
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
	err := os.Remove(f.FilePath(id))
	f.unlockId(id)
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
