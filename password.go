package password

// Managers stores a map of string identifiers for all created password managers.
// The identifier "default" always holds the default manager from GetDefaultManager.
// It can be set via SetDefaultManager. Do not manipulate directly.
var Managers map[string]*Manager = map[string]*Manager{
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

// ToggleHashPassword will toggle the config variable HashPassword of the default password manager and return the current state.
func ToggleHashPassword() bool {
	m := GetDefaultManager()
	m.HashPassword = !m.HashPassword
	return m.HashPassword
}

// EnableRecovery will enforce recovery key file storage alongside passwords.
func EnableRecovery(key string) {
	GetDefaultManager().EnableRecovery(key)
}

// DisableRecovery will stop recovery key file storage alongside passwords.
func DisableRecovery() {
	GetDefaultManager().DisableRecovery()
}

// SetTemporaryStorage overwrites the current storage backend with a memory based one.
func SetTemporaryStorage() {
	GetDefaultManager().storageBackend = NewTemporaryStorage()
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
