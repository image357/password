package password

import (
	"fmt"
	"github.com/image357/password/log"
	"strings"
)

// RecoveryIdSuffix stores the id and file suffix that identifies recovery key files.
const RecoveryIdSuffix string = ".recovery"

type Manager struct {
	// HashPassword signals if passwords will be stored as hashes.
	HashPassword bool

	// withRecovery signals that a recovery key file must be stored alongside passwords.
	withRecovery bool

	// recoveryKeyBytes store the result of EncryptOTP such that the recovery key is obfuscated in memory.
	recoveryKeyBytes []byte
	// recoveryKeySecret store the result of EncryptOTP such that the recovery key is obfuscated in memory.
	recoveryKeySecret []byte

	// storageBackend handles password storage.
	storageBackend Storage
}

// NewManager creates a new passwordManager instance and applies basic initialization.
func NewManager() *Manager {
	m := new(Manager)

	m.HashPassword = false
	m.withRecovery = false
	m.storageBackend = NewFileStorage()

	return m
}

// EnableRecovery will enforce recovery key file storage alongside passwords.
func (m *Manager) EnableRecovery(key string) {
	m.withRecovery = true
	m.recoveryKeyBytes, m.recoveryKeySecret = EncryptOTP(key)
}

// DisableRecovery will stop recovery key file storage alongside passwords.
func (m *Manager) DisableRecovery() {
	m.withRecovery = false
	m.recoveryKeyBytes, m.recoveryKeySecret = nil, nil
}

// getRecoveryKey returns the recovery key that was set by EnableRecovery.
func (m *Manager) getRecoveryKey() string {
	return DecryptOTP(m.recoveryKeyBytes, m.recoveryKeySecret)
}

// Overwrite an existing password or create a new one.
// key is the encryption secret for storage.
func (m *Manager) Overwrite(id string, password string, key string) error {
	id = NormalizeId(id)

	if m.HashPassword && !(m.withRecovery && strings.HasSuffix(id, RecoveryIdSuffix)) {
		hashedPassword, err := getHashedPassword(password)
		if err != nil {
			return err
		}
		password = hashedPassword
	}

	packedData, err := packData(id, password)
	if err != nil {
		return err
	}

	encryptedData, err := Encrypt(packedData, key)
	if err != nil {
		return err
	}

	err = m.storageBackend.Store(id, encryptedData)
	if err != nil {
		return err
	}

	if m.withRecovery && !strings.HasSuffix(id, RecoveryIdSuffix) {
		// write recovery key file
		recoveryId := id + RecoveryIdSuffix
		err = m.Overwrite(recoveryId, key, m.getRecoveryKey())
		if err != nil {
			log.Warn("cannot write recovery key file", "id", recoveryId)
		}
	}

	return nil
}

// Get an existing password with id.
// key is the encryption secret for storage.
func (m *Manager) Get(id string, key string) (string, error) {
	id = NormalizeId(id)

	encryptedData, err := m.storageBackend.Retrieve(id)
	if err != nil {
		return "", err
	}

	packedData, err := Decrypt(encryptedData, key)
	if err != nil {
		return "", err
	}

	storedId, password, err := unpackData(packedData)
	if err != nil {
		return "", err
	}
	if storedId != id {
		return "", fmt.Errorf("storage id mismatch")
	}

	return password, nil
}

// Check an existing password for equality with the provided password.
// key is the encryption secret for storage.
func (m *Manager) Check(id string, password string, key string) (bool, error) {
	id = NormalizeId(id)

	decryptedPassword, err := m.Get(id, key)
	if err != nil {
		return false, err
	}

	var result bool
	if m.HashPassword && !(m.withRecovery && strings.HasSuffix(id, RecoveryIdSuffix)) {
		result, err = compareHashedPassword(decryptedPassword, password)
		if err != nil {
			return false, err
		}
	} else {
		result = comparePassword(decryptedPassword, password)
	}

	return result, nil
}

// Set an existing password-id or create a new one.
// oldPassword must match the currently stored password.
// key is the encryption secret for storage.
func (m *Manager) Set(id string, oldPassword string, newPassword string, key string) error {
	id = NormalizeId(id)

	exists, err := m.storageBackend.Exists(id)
	if err != nil {
		return err
	}

	if exists {
		correct, err := m.Check(id, oldPassword, key)
		if err != nil {
			return err
		}
		if !correct {
			return fmt.Errorf("password is incorrect")
		}
	}

	err = m.Overwrite(id, newPassword, key)
	if err != nil {
		return err
	}
	return nil
}

// Unset (delete) an existing password.
// password must match the currently stored password.
// key is the encryption secret for storage.
func (m *Manager) Unset(id string, password string, key string) error {
	id = NormalizeId(id)

	correct, err := m.Check(id, password, key)
	if err != nil {
		return err
	}
	if !correct {
		return fmt.Errorf("password is incorrect")
	}

	return m.storageBackend.Delete(id)
}

// Exists tests if a given id already exists in the storage backend.
func (m *Manager) Exists(id string) (bool, error) {
	id = NormalizeId(id)
	return m.storageBackend.Exists(id)
}

// List all stored password-ids.
func (m *Manager) List() ([]string, error) {
	return m.storageBackend.List()
}

// Delete an existing password.
func (m *Manager) Delete(id string) error {
	id = NormalizeId(id)
	return m.storageBackend.Delete(id)
}

// Clean (delete) all stored passwords.
func (m *Manager) Clean() error {
	return m.storageBackend.Clean()
}

// RewriteKey changes the storage key of a password file from oldKey to newKey.
// Encryption hashes will be renewed, stored metadata will be unchanged.
// If enabled, recovery files will be recreated.
func (m *Manager) RewriteKey(id string, oldKey string, newKey string) error {
	id = NormalizeId(id)

	encryptedData, err := m.storageBackend.Retrieve(id)
	if err != nil {
		return err
	}

	packedData, err := Decrypt(encryptedData, oldKey)
	if err != nil {
		return err
	}

	newData, err := Encrypt(packedData, newKey)
	if err != nil {
		return err
	}

	err = m.storageBackend.Store(id, newData)
	if err != nil {
		return err
	}

	if m.withRecovery && !strings.HasSuffix(id, RecoveryIdSuffix) {
		// write recovery key file
		recoveryId := id + RecoveryIdSuffix
		err = m.Overwrite(recoveryId, newKey, m.getRecoveryKey())
		if err != nil {
			log.Warn("cannot write recovery key file", "id", recoveryId)
		}
	}

	return nil
}
