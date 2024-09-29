package password

import (
	"fmt"
	"os"
)

// Overwrite an existing password or create a new one.
// key is the encryption secret for storage.
func Overwrite(id string, password string, key string) error {
	id = NormalizeId(id)

	encryptedPassword, err := encrypt(password, id+key)
	if err != nil {
		return err
	}

	err = store(id, encryptedPassword)
	if err != nil {
		return err
	}

	return nil
}

// Get an existing password with id.
// key is the encryption secret for storage.
func Get(id string, key string) (string, error) {
	id = NormalizeId(id)

	encryptedPassword, err := retrieve(id)
	if err != nil {
		return "", err
	}

	decryptedPassword, err := decrypt(encryptedPassword, id+key)
	if err != nil {
		return "", err
	}

	return decryptedPassword, nil
}

// Check an existing password for equality with the provided password.
// key is the encryption secret for storage.
func Check(id string, password string, key string) (bool, error) {
	decryptedPassword, err := Get(id, key)
	if err != nil {
		return false, err
	}

	if decryptedPassword == password {
		return true, nil
	}
	return false, nil
}

// Set an existing password-id or create a new one.
// oldPassword must match the currently stored password.
// key is the encryption secret for storage.
func Set(id string, oldPassword string, newPassword string, key string) error {
	if _, err := os.Stat(FilePath(id)); !os.IsNotExist(err) {
		correct, err := Check(id, oldPassword, key)
		if err != nil {
			return err
		}
		if !correct {
			return fmt.Errorf("password is incorrect")
		}
	}

	err := Overwrite(id, newPassword, key)
	if err != nil {
		return err
	}
	return nil
}

// Unset (delete) an existing password.
// password must match the currently stored password.
// key is the encryption secret for storage.
func Unset(id string, password string, key string) error {
	correct, err := Check(id, password, key)
	if err != nil {
		return err
	}
	if !correct {
		return fmt.Errorf("password is incorrect")
	}

	return Delete(id)
}
