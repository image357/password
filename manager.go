package password

import (
	"fmt"
	"github.com/image357/password/log"
	"strings"
)

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
	pm := new(Manager)

	pm.HashPassword = false
	pm.withRecovery = false
	pm.storageBackend = newFileStorage()

	return pm
}

// EnableRecovery will enforce recovery key file storage alongside passwords.
func (pm *Manager) EnableRecovery(key string) {
	pm.withRecovery = true
	pm.recoveryKeyBytes, pm.recoveryKeySecret = EncryptOTP(key)
}

// DisableRecovery will stop recovery key file storage alongside passwords.
func (pm *Manager) DisableRecovery() {
	pm.withRecovery = false
	pm.recoveryKeyBytes, pm.recoveryKeySecret = nil, nil
}

// getRecoveryKey returns the recovery key that was set by EnableRecovery.
func (pm *Manager) getRecoveryKey() string {
	return DecryptOTP(pm.recoveryKeyBytes, pm.recoveryKeySecret)
}

// Overwrite an existing password or create a new one.
// key is the encryption secret for storage.
func (pm *Manager) Overwrite(id string, password string, key string) error {
	id = NormalizeId(id)

	if pm.HashPassword && !(pm.withRecovery && strings.HasSuffix(id, RecoveryIdSuffix)) {
		hashedPassword, err := getHashedPassword(password)
		if err != nil {
			return err
		}
		password = hashedPassword
	}

	data, err := packData(id, password)
	if err != nil {
		return err
	}

	encryptedData, err := encrypt(data, key)
	if err != nil {
		return err
	}

	err = pm.storageBackend.Store(id, encryptedData)
	if err != nil {
		return err
	}

	if pm.withRecovery && !strings.HasSuffix(id, RecoveryIdSuffix) {
		// write recovery key file
		recoveryId := id + RecoveryIdSuffix
		err = pm.Overwrite(recoveryId, key, pm.getRecoveryKey())
		if err != nil {
			log.Warn("cannot write recovery key file", "id", recoveryId)
		}
	}

	return nil
}

// Get an existing password with id.
// key is the encryption secret for storage.
func (pm *Manager) Get(id string, key string) (string, error) {
	id = NormalizeId(id)

	encryptedData, err := pm.storageBackend.Retrieve(id)
	if err != nil {
		return "", err
	}

	decryptedData, err := decrypt(encryptedData, key)
	if err != nil {
		return "", err
	}

	storedId, password, err := unpackData(decryptedData)
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
func (pm *Manager) Check(id string, password string, key string) (bool, error) {
	decryptedPassword, err := pm.Get(id, key)
	if err != nil {
		return false, err
	}

	var result bool
	if pm.HashPassword && !(pm.withRecovery && strings.HasSuffix(id, RecoveryIdSuffix)) {
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
func (pm *Manager) Set(id string, oldPassword string, newPassword string, key string) error {
	exists, err := pm.storageBackend.Exists(id)
	if err != nil {
		return err
	}

	if exists {
		correct, err := pm.Check(id, oldPassword, key)
		if err != nil {
			return err
		}
		if !correct {
			return fmt.Errorf("password is incorrect")
		}
	}

	err = pm.Overwrite(id, newPassword, key)
	if err != nil {
		return err
	}
	return nil
}

// Unset (delete) an existing password.
// password must match the currently stored password.
// key is the encryption secret for storage.
func (pm *Manager) Unset(id string, password string, key string) error {
	correct, err := pm.Check(id, password, key)
	if err != nil {
		return err
	}
	if !correct {
		return fmt.Errorf("password is incorrect")
	}

	return pm.storageBackend.Delete(id)
}
