// Package password provides a simple-password-manager library with an encryption backend to handle app passwords.
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
package password

// Managers stores a map of string identifiers for all created password managers.
// The identifier "default" always holds the default manager from GetDefaultManager.
// It can be set via SetDefaultManager. Do not manipulate directly.
var Managers = map[string]*Manager{
	"default": NewManager(),
}

// GetDefaultManager returns the current default password manager.
func GetDefaultManager() *Manager {
	return Managers["default"]
}

// SetDefaultManager will overwrite the current default password manager with the provided one.
func SetDefaultManager(manager *Manager) {
	Managers["default"] = manager
}

// RegisterDefaultManager will register the current default password manger under the identifier and set a new default manager.
func RegisterDefaultManager(identifier string) {
	Managers[identifier] = GetDefaultManager()
	SetDefaultManager(NewManager())
}

// EnableHashing will set the config variable Manager.HashPassword of the default password manager to true.
// This enables storage of hashed passwords.
func EnableHashing() {
	m := GetDefaultManager()
	m.HashPassword = true
}

// DisableHashing will set the config variable Manager.HashPassword of the default password manager to false.
// This disables storage of hashed passwords.
func DisableHashing() {
	m := GetDefaultManager()
	m.HashPassword = false
}

// EnableRecovery will enforce recovery key file storage alongside passwords.
func EnableRecovery(key string) {
	GetDefaultManager().EnableRecovery(key)
}

// DisableRecovery will stop recovery key file storage alongside passwords.
func DisableRecovery() {
	GetDefaultManager().DisableRecovery()
}

// Overwrite an existing password or create a new one.
// key is the encryption secret for storage.
func Overwrite(id string, password string, key string) error {
	return GetDefaultManager().Overwrite(id, password, key)
}

// Get an existing password with id.
// key is the encryption secret for storage.
func Get(id string, key string) (string, error) {
	return GetDefaultManager().Get(id, key)
}

// Check an existing password for equality with the provided password.
// key is the encryption secret for storage.
func Check(id string, password string, key string) (bool, error) {
	return GetDefaultManager().Check(id, password, key)
}

// Set an existing password-id or create a new one.
// oldPassword must match the currently stored password.
// key is the encryption secret for storage.
func Set(id string, oldPassword string, newPassword string, key string) error {
	return GetDefaultManager().Set(id, oldPassword, newPassword, key)
}

// Unset (delete) an existing password.
// password must match the currently stored password.
// key is the encryption secret for storage.
func Unset(id string, password string, key string) error {
	return GetDefaultManager().Unset(id, password, key)
}

// Exists tests if a given id already exists in the storage backend.
func Exists(id string) (bool, error) {
	return GetDefaultManager().Exists(id)
}

// List all stored password-ids.
func List() ([]string, error) {
	return GetDefaultManager().List()
}

// Delete an existing password.
func Delete(id string) error {
	return GetDefaultManager().Delete(id)
}

// Clean (delete) all stored passwords.
func Clean() error {
	return GetDefaultManager().Clean()
}

// RewriteKey changes the storage key of a password from oldKey to newKey.
// Encryption hashes will be renewed. Stored metadata will be unchanged.
// If enabled, recovery entries will be recreated.
func RewriteKey(id string, oldKey string, newKey string) error {
	return GetDefaultManager().RewriteKey(id, oldKey, newKey)
}
