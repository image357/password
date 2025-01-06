package main

import (
	pwd "github.com/image357/password"
	"github.com/image357/password/log"
	"github.com/image357/password/rest"
	"log/slog"
	"strings"
	"unsafe"
)

/*

#ifdef __cplusplus
#define _Bool bool
#else
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#endif

#define CPWD__LevelError  8
#define CPWD__LevelWarn   4
#define CPWD__LevelInfo   0
#define CPWD__LevelDebug -4

typedef const char cchar_t;
typedef bool (*CPWD__TestAccessFunc)(cchar_t *token, cchar_t *ip, cchar_t *resource, cchar_t *id);

static bool CPWD__RunCallback(cchar_t *token, cchar_t *ip, cchar_t *resource, cchar_t *id, CPWD__TestAccessFunc callback) {
	return callback(token, ip, resource, id);
}
*/
import "C"

func main() {}

// CPWD__SetDefaultManager calls password.SetDefaultManager.
// Returns 0 if the identifier has already been registered, -1 if not.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__SetDefaultManager
func CPWD__SetDefaultManager(identifier *C.cchar_t) int {
	i := C.GoString(identifier)
	m, ok := pwd.Managers[i]
	if !ok {
		delete(pwd.Managers, i)
		return -1
	}
	pwd.SetDefaultManager(m)
	return 0
}

// CPWD__RegisterDefaultManager calls password.RegisterDefaultManager.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__RegisterDefaultManager
func CPWD__RegisterDefaultManager(identifier *C.cchar_t) {
	i := C.GoString(identifier)
	pwd.RegisterDefaultManager(i)
}

// CPWD__EnableHashing calls password.EnableHashing.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__EnableHashing
func CPWD__EnableHashing() {
	pwd.EnableHashing()
}

// CPWD__DisableHashing calls password.DisableHashing.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__DisableHashing
func CPWD__DisableHashing() {
	pwd.DisableHashing()
}

// CPWD__EnableRecovery calls password.EnableRecovery.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__EnableRecovery
func CPWD__EnableRecovery(key *C.cchar_t) {
	pwd.EnableRecovery(C.GoString(key))
}

// CPWD__DisableRecovery calls password.DisableRecovery.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__DisableRecovery
func CPWD__DisableRecovery() {
	pwd.DisableRecovery()
}

// CPWD__Overwrite calls password.Overwrite and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__Overwrite
func CPWD__Overwrite(id *C.cchar_t, password *C.cchar_t, key *C.cchar_t) int {
	err := pwd.Overwrite(C.GoString(id), C.GoString(password), C.GoString(key))
	if err != nil {
		log.Error("CPWD__Overwrite: Overwrite failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__mOverwrite calls password.Overwrite with the specified manager and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__mOverwrite
func CPWD__mOverwrite(manager *C.cchar_t, id *C.cchar_t, password *C.cchar_t, key *C.cchar_t) int {
	identifier := C.GoString(manager)
	m, ok := pwd.Managers[identifier]
	if !ok {
		delete(pwd.Managers, identifier)
		log.Error("CPWD__mOverwrite: Manager not found", "identifier", identifier)
		return -1
	}

	err := m.Overwrite(C.GoString(id), C.GoString(password), C.GoString(key))
	if err != nil {
		log.Error("CPWD__mOverwrite: Overwrite failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__Get calls password.Get and returns 0 on success, -1 on error.
// The result will be stored in buffer.
//
// For full documentation visit https://github.com/image357/password/docs/password.md
//
//export CPWD__Get
func CPWD__Get(id *C.cchar_t, key *C.cchar_t, buffer *C.char, length int) int {
	if buffer == nil {
		log.Error("CPWD__Get: buffer is nullptr")
		return -1
	}

	password, err := pwd.Get(C.GoString(id), C.GoString(key))
	if err != nil {
		log.Error("CPWD__Get: Get failed", "error", err)
		return -1
	}

	cs := C.CString(password)
	defer C.free(unsafe.Pointer(cs))
	if int(C.strlen(cs)) >= length {
		log.Error("CPWD__Get: buffer is too small")
		return -1
	}
	C.strcpy(buffer, cs)

	return 0
}

// CPWD__mGet calls password.Get with the specified manager and returns 0 on success, -1 on error.
// The result will be stored in buffer.
//
// For full documentation visit https://github.com/image357/password/docs/password.md
//
//export CPWD__mGet
func CPWD__mGet(manager *C.cchar_t, id *C.cchar_t, key *C.cchar_t, buffer *C.char, length int) int {
	identifier := C.GoString(manager)
	m, ok := pwd.Managers[identifier]
	if !ok {
		delete(pwd.Managers, identifier)
		log.Error("CPWD__mGet: Manager not found", "identifier", identifier)
		return -1
	}

	if buffer == nil {
		log.Error("CPWD__mGet: buffer is nullptr")
		return -1
	}

	password, err := m.Get(C.GoString(id), C.GoString(key))
	if err != nil {
		log.Error("CPWD__mGet: Get failed", "error", err)
		return -1
	}

	cs := C.CString(password)
	defer C.free(unsafe.Pointer(cs))
	if int(C.strlen(cs)) >= length {
		log.Error("CPWD__mGet: buffer is too small")
		return -1
	}
	C.strcpy(buffer, cs)

	return 0
}

// CPWD__Check calls password.Check and returns 0 on success, -1 on error.
// The result will be stored via the result pointer.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__Check
func CPWD__Check(id *C.cchar_t, password *C.cchar_t, key *C.cchar_t, result *C.bool) int {
	if result == nil {
		log.Error("CPWD__Check: result is nullptr")
		return -1
	}

	check, err := pwd.Check(C.GoString(id), C.GoString(password), C.GoString(key))
	if err != nil {
		log.Error("CPWD__Check: Check failed", "error", err)
		return -1
	}

	*result = C.bool(check)
	return 0
}

// CPWD__mCheck calls password.Check with the specified manager and returns 0 on success, -1 on error.
// The result will be stored via the result pointer.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__mCheck
func CPWD__mCheck(manager *C.cchar_t, id *C.cchar_t, password *C.cchar_t, key *C.cchar_t, result *C.bool) int {
	identifier := C.GoString(manager)
	m, ok := pwd.Managers[identifier]
	if !ok {
		delete(pwd.Managers, identifier)
		log.Error("CPWD__mCheck: Manager not found", "identifier", identifier)
		return -1
	}

	if result == nil {
		log.Error("CPWD__mCheck: result is nullptr")
		return -1
	}

	check, err := m.Check(C.GoString(id), C.GoString(password), C.GoString(key))
	if err != nil {
		log.Error("CPWD__mCheck: Check failed", "error", err)
		return -1
	}

	*result = C.bool(check)
	return 0
}

// CPWD__Set calls password.Set and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__Set
func CPWD__Set(id *C.cchar_t, oldPassword *C.cchar_t, newPassword *C.cchar_t, key *C.cchar_t) int {
	err := pwd.Set(C.GoString(id), C.GoString(oldPassword), C.GoString(newPassword), C.GoString(key))
	if err != nil {
		log.Error("CPWD__Set: Set failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__mSet calls password.Set with the specified manager and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__mSet
func CPWD__mSet(manager *C.cchar_t, id *C.cchar_t, oldPassword *C.cchar_t, newPassword *C.cchar_t, key *C.cchar_t) int {
	identifier := C.GoString(manager)
	m, ok := pwd.Managers[identifier]
	if !ok {
		delete(pwd.Managers, identifier)
		log.Error("CPWD__mSet: Manager not found", "identifier", identifier)
		return -1
	}

	err := m.Set(C.GoString(id), C.GoString(oldPassword), C.GoString(newPassword), C.GoString(key))
	if err != nil {
		log.Error("CPWD__mSet: Set failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__Unset calls password.Unset and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__Unset
func CPWD__Unset(id *C.cchar_t, password *C.cchar_t, key *C.cchar_t) int {
	err := pwd.Unset(C.GoString(id), C.GoString(password), C.GoString(key))
	if err != nil {
		log.Error("CPWD__Unset: Unset failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__mUnset calls password.Unset with the specified manager and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__mUnset
func CPWD__mUnset(manager *C.cchar_t, id *C.cchar_t, password *C.cchar_t, key *C.cchar_t) int {
	identifier := C.GoString(manager)
	m, ok := pwd.Managers[identifier]
	if !ok {
		delete(pwd.Managers, identifier)
		log.Error("CPWD__mUnset: Manager not found", "identifier", identifier)
		return -1
	}

	err := m.Unset(C.GoString(id), C.GoString(password), C.GoString(key))
	if err != nil {
		log.Error("CPWD__mUnset: Unset failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__Exists calls password.Exists and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__Exists
func CPWD__Exists(id *C.cchar_t, result *C.bool) int {
	if result == nil {
		log.Error("CPWD__Exists: result is nullptr")
		return -1
	}

	exists, err := pwd.Exists(C.GoString(id))
	if err != nil {
		log.Error("CPWD__Exists: Exists failed", "error", err)
		return -1
	}

	*result = C.bool(exists)
	return 0
}

// CPWD__mExists calls password.Exists with the specified manager and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__mExists
func CPWD__mExists(manager *C.cchar_t, id *C.cchar_t, result *C.bool) int {
	identifier := C.GoString(manager)
	m, ok := pwd.Managers[identifier]
	if !ok {
		delete(pwd.Managers, identifier)
		log.Error("CPWD__mExists: Manager not found", "identifier", identifier)
		return -1
	}

	if result == nil {
		log.Error("CPWD__mExists: result is nullptr")
		return -1
	}

	exists, err := m.Exists(C.GoString(id))
	if err != nil {
		log.Error("CPWD__mExists: Exists failed", "error", err)
		return -1
	}

	*result = C.bool(exists)
	return 0
}

// CPWD__List calls password.List and returns 0 on success, -1 on error.
// The resulting list will be stored in buffer with delim as separator.
// Error is returned if delim collides with any of the returned ids.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__List
func CPWD__List(buffer *C.char, length int, delim *C.cchar_t) int {
	if buffer == nil {
		log.Error("CPWD__List: buffer is nullptr")
		return -1
	}

	list, err := pwd.List()
	if err != nil {
		log.Error("CPWD__List: List failed", "error", err)
		return -1
	}

	d := C.GoString(delim)
	for _, l := range list {
		if strings.Contains(l, d) {
			log.Error("CPWD__List: delimiter collision with id", "delim", d, "id", l)
			return -1
		}
	}
	s := strings.Join(list, d)

	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if int(C.strlen(cs)) >= length {
		log.Error("CPWD__List: buffer is too small")
		return -1
	}
	C.strcpy(buffer, cs)

	return 0
}

// CPWD__mList calls password.List with the specified manager and returns 0 on success, -1 on error.
// The resulting list will be stored in buffer with delim as separator.
// Error is returned if delim collides with any of the returned ids.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__mList
func CPWD__mList(manager *C.cchar_t, buffer *C.char, length int, delim *C.cchar_t) int {
	identifier := C.GoString(manager)
	m, ok := pwd.Managers[identifier]
	if !ok {
		delete(pwd.Managers, identifier)
		log.Error("CPWD__mList: Manager not found", "identifier", identifier)
		return -1
	}

	if buffer == nil {
		log.Error("CPWD__mList: buffer is nullptr")
		return -1
	}

	list, err := m.List()
	if err != nil {
		log.Error("CPWD__mList: List failed", "error", err)
		return -1
	}

	d := C.GoString(delim)
	for _, l := range list {
		if strings.Contains(l, d) {
			log.Error("CPWD__mList: delimiter collision with id", "delim", d, "id", l)
			return -1
		}
	}
	s := strings.Join(list, d)

	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if int(C.strlen(cs)) >= length {
		log.Error("CPWD__mList: buffer is too small")
		return -1
	}
	C.strcpy(buffer, cs)

	return 0
}

// CPWD__Delete calls password.Delete and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__Delete
func CPWD__Delete(id *C.cchar_t) int {
	err := pwd.Delete(C.GoString(id))
	if err != nil {
		log.Error("CPWD__Delete: Delete failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__mDelete calls password.Delete with the specified manager and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__mDelete
func CPWD__mDelete(manager *C.cchar_t, id *C.cchar_t) int {
	identifier := C.GoString(manager)
	m, ok := pwd.Managers[identifier]
	if !ok {
		delete(pwd.Managers, identifier)
		log.Error("CPWD__mDelete: Manager not found", "identifier", identifier)
		return -1
	}

	err := m.Delete(C.GoString(id))
	if err != nil {
		log.Error("CPWD__mDelete: Delete failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__Clean calls password.Clean and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__Clean
func CPWD__Clean() int {
	err := pwd.Clean()
	if err != nil {
		log.Error("CPWD__Clean: Clean failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__mClean calls password.Clean with the specified manager and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__mClean
func CPWD__mClean(manager *C.cchar_t) int {
	identifier := C.GoString(manager)
	m, ok := pwd.Managers[identifier]
	if !ok {
		delete(pwd.Managers, identifier)
		log.Error("CPWD__mClean: Manager not found", "identifier", identifier)
		return -1
	}

	err := m.Clean()
	if err != nil {
		log.Error("CPWD__mClean: Clean failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__RewriteKey calls password.RewriteKey and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__RewriteKey
func CPWD__RewriteKey(id *C.cchar_t, oldKey *C.cchar_t, newKey *C.cchar_t) int {
	err := pwd.RewriteKey(C.GoString(id), C.GoString(oldKey), C.GoString(newKey))
	if err != nil {
		log.Error("CPWD__RewriteKey: RewriteKey failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__mRewriteKey calls password.RewriteKey with the specified manager and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__mRewriteKey
func CPWD__mRewriteKey(manager *C.cchar_t, id *C.cchar_t, oldKey *C.cchar_t, newKey *C.cchar_t) int {
	identifier := C.GoString(manager)
	m, ok := pwd.Managers[identifier]
	if !ok {
		delete(pwd.Managers, identifier)
		log.Error("CPWD__mRewriteKey: Manager not found", "identifier", identifier)
		return -1
	}

	err := m.RewriteKey(C.GoString(id), C.GoString(oldKey), C.GoString(newKey))
	if err != nil {
		log.Error("CPWD__mRewriteKey: RewriteKey failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__StartSimpleService calls rest.StartSimpleService and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/rest.md
//
//export CPWD__StartSimpleService
func CPWD__StartSimpleService(bindAddress *C.cchar_t, prefix *C.cchar_t, key *C.cchar_t, callback C.CPWD__TestAccessFunc) int {
	if callback == nil {
		log.Error("CPWD__StartSimpleService: callback is nullptr")
		return -1
	}

	err := rest.StartSimpleService(C.GoString(bindAddress), C.GoString(prefix), C.GoString(key),
		func(token string, ip string, resource string, id string) bool {
			cToken := C.CString(token)
			defer C.free(unsafe.Pointer(cToken))

			cIP := C.CString(ip)
			defer C.free(unsafe.Pointer(cIP))

			cResource := C.CString(resource)
			defer C.free(unsafe.Pointer(cResource))

			cId := C.CString(id)
			defer C.free(unsafe.Pointer(cId))

			value := C.CPWD__RunCallback(cToken, cIP, cResource, cId, callback)
			return bool(value)
		})
	if err != nil {
		log.Error("CPWD__StartSimpleService: rest.StartSimpleService failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__StartMultiService calls rest.StartMultiService and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/rest.md
//
//export CPWD__StartMultiService
func CPWD__StartMultiService(bindAddress *C.cchar_t, prefix *C.cchar_t, key *C.cchar_t, callback C.CPWD__TestAccessFunc) int {
	if callback == nil {
		log.Error("CPWD__StartMultiService: callback is nullptr")
		return -1
	}

	err := rest.StartMultiService(C.GoString(bindAddress), C.GoString(prefix), C.GoString(key),
		func(token string, ip string, resource string, id string) bool {
			cToken := C.CString(token)
			defer C.free(unsafe.Pointer(cToken))

			cIP := C.CString(ip)
			defer C.free(unsafe.Pointer(cIP))

			cResource := C.CString(resource)
			defer C.free(unsafe.Pointer(cResource))

			cId := C.CString(id)
			defer C.free(unsafe.Pointer(cId))

			value := C.CPWD__RunCallback(cToken, cIP, cResource, cId, callback)
			return bool(value)
		})
	if err != nil {
		log.Error("CPWD__StartMultiService: rest.StartMultiService failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__StopService calls rest.StopService and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/rest.md
//
//export CPWD__StopService
func CPWD__StopService(timeout int, bindAddress *C.cchar_t, prefix *C.cchar_t) int {
	err := rest.StopService(timeout, C.GoString(bindAddress), C.GoString(prefix))
	if err != nil {
		log.Error("CPWD__StopService: rest.StopService failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__NormalizeId calls password.NormalizeId and returns 0 on success, -1 on error.
// The result will be stored in buffer.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__NormalizeId
func CPWD__NormalizeId(id *C.cchar_t, buffer *C.char, length int) int {
	if buffer == nil {
		log.Error("CPWD__NormalizeId: buffer is nullptr")
		return -1
	}

	s := pwd.NormalizeId(C.GoString(id))

	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if int(C.strlen(cs)) >= length {
		log.Error("CPWD__GetStorePath: buffer is too small")
		return -1
	}
	C.strcpy(buffer, cs)

	return 0
}

// CPWD__GetStorePath calls password.GetStorePath and returns 0 on success, -1 on error.
// The result will be stored in buffer.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__GetStorePath
func CPWD__GetStorePath(buffer *C.char, length int) int {
	if buffer == nil {
		log.Error("CPWD__GetStorePath: buffer is nullptr")
		return -1
	}

	s, err := pwd.GetStorePath()
	if err != nil {
		log.Error("CPWD__GetStorePath: GetStorePath failed", "error", err)
		return -1
	}

	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if int(C.strlen(cs)) >= length {
		log.Error("CPWD__GetStorePath: buffer is too small")
		return -1
	}
	C.strcpy(buffer, cs)

	return 0
}

// CPWD__SetStorePath calls password.SetStorePath and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__SetStorePath
func CPWD__SetStorePath(path *C.cchar_t) int {
	err := pwd.SetStorePath(C.GoString(path))
	if err != nil {
		log.Error("CPWD__SetStorePath: SetStorePath failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__GetFileEnding calls password.GetFileEnding and returns 0 on success, -1 on error.
// The result will be stored in buffer.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__GetFileEnding
func CPWD__GetFileEnding(buffer *C.char, length int) int {
	if buffer == nil {
		log.Error("CPWD__GetFileEnding: buffer is nullptr")
		return -1
	}

	s, err := pwd.GetFileEnding()
	if err != nil {
		log.Error("CPWD__GetFileEnding: GetFileEnding failed", "error", err)
		return -1
	}

	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if int(C.strlen(cs)) >= length {
		log.Error("CPWD__GetFileEnding: buffer is too small")
		return -1
	}
	C.strcpy(buffer, cs)

	return 0
}

// CPWD__SetFileEnding calls password.SetFileEnding and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__SetFileEnding
func CPWD__SetFileEnding(e *C.cchar_t) int {
	err := pwd.SetFileEnding(C.GoString(e))
	if err != nil {
		log.Error("CPWD__SetFileEnding: SetFileEnding failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__FilePath calls password.FilePath and returns 0 on success, -1 on error.
// The result will be stored in buffer.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__FilePath
func CPWD__FilePath(id *C.cchar_t, buffer *C.char, length int) int {
	if buffer == nil {
		log.Error("CPWD__FilePath: buffer is nullptr")
		return -1
	}

	s, err := pwd.FilePath(C.GoString(id))
	if err != nil {
		log.Error("CPWD__FilePath: FilePath failed", "error", err)
		return -1
	}

	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if int(C.strlen(cs)) >= length {
		log.Error("CPWD__FilePath: buffer is too small")
		return -1
	}
	C.strcpy(buffer, cs)

	return 0
}

// CPWD__SetTemporaryStorage calls password.SetTemporaryStorage.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__SetTemporaryStorage
func CPWD__SetTemporaryStorage() {
	pwd.SetTemporaryStorage()
}

// CPWD__DumpJSON calls password.DumpJSON and returns 0 on success, -1 on error.
// The result will be stored in buffer.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__DumpJSON
func CPWD__DumpJSON(buffer *C.char, length int) int {
	if buffer == nil {
		log.Error("CPWD__DumpJSON: buffer is nullptr")
		return -1
	}

	s, err := pwd.DumpJSON()
	if err != nil {
		log.Error("CPWD__DumpJSON: DumpJSON failed", "error", err)
		return -1
	}

	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if int(C.strlen(cs)) >= length {
		log.Error("CPWD__DumpJSON: buffer is too small")
		return -1
	}
	C.strcpy(buffer, cs)

	return 0
}

// CPWD__LoadJSON calls password.LoadJSON and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__LoadJSON
func CPWD__LoadJSON(input *C.cchar_t) int {
	err := pwd.LoadJSON(C.GoString(input))
	if err != nil {
		log.Error("CPWD__LoadJSON: LoadJSON failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__WriteToDisk calls password.WriteToDisk and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__WriteToDisk
func CPWD__WriteToDisk(path *C.cchar_t) int {
	err := pwd.WriteToDisk(C.GoString(path))
	if err != nil {
		log.Error("CPWD__WriteToDisk: WriteToDisk failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__ReadFromDisk calls password.ReadFromDisk and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/password.md
//
//export CPWD__ReadFromDisk
func CPWD__ReadFromDisk(path *C.cchar_t) int {
	err := pwd.ReadFromDisk(C.GoString(path))
	if err != nil {
		log.Error("CPWD__ReadFromDisk: ReadFromDisk failed", "error", err)
		return -1
	}
	return 0
}

// CPWD__LogLevel calls log.Level and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/log.md
//
//export CPWD__LogLevel
func CPWD__LogLevel(level int) int {
	switch level {
	case C.CPWD__LevelError:
		log.Level(slog.LevelError)
	case C.CPWD__LevelWarn:
		log.Level(slog.LevelWarn)
	case C.CPWD__LevelInfo:
		log.Level(slog.LevelInfo)
	case C.CPWD__LevelDebug:
		log.Level(slog.LevelDebug)
	default:
		log.Error("CPWD__LogLevel: unknown level")
		return -1
	}
	return 0
}

// CPWD__LogSetDefault calls log.SetDefault.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/log.md
//
//export CPWD__LogSetDefault
func CPWD__LogSetDefault() {
	log.SetDefault()
}

// CPWD__LogSetStderrText calls log.SetStderrText.
//
//export CPWD__LogSetStderrText
func CPWD__LogSetStderrText() {
	log.SetStderrText()
}

// CPWD__LogSetStderrJSON calls log.SetStderrJSON.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/log.md
//
//export CPWD__LogSetStderrJSON
func CPWD__LogSetStderrJSON() {
	log.SetStderrJSON()
}

// CPWD__LogSetFileText calls log.SetFileText and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/log.md
//
//export CPWD__LogSetFileText
func CPWD__LogSetFileText(filePath *C.cchar_t) int {
	err := log.SetFileText(C.GoString(filePath))
	if err != nil {
		return -1
	}
	return 0
}

// CPWD__LogSetFileJSON calls log.SetFileJSON and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/log.md
//
//export CPWD__LogSetFileJSON
func CPWD__LogSetFileJSON(filePath *C.cchar_t) int {
	err := log.SetFileJSON(C.GoString(filePath))
	if err != nil {
		return -1
	}
	return 0
}

// CPWD__LogSetMultiText calls log.SetMultiText and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/log.md
//
//export CPWD__LogSetMultiText
func CPWD__LogSetMultiText(filePath *C.cchar_t) int {
	err := log.SetMultiText(C.GoString(filePath))
	if err != nil {
		return -1
	}
	return 0
}

// CPWD__LogSetMultiJSON calls log.SetMultiJSON and returns 0 on success, -1 on error.
//
// For full documentation visit https://github.com/image357/password/blob/main/docs/log.md
//
//export CPWD__LogSetMultiJSON
func CPWD__LogSetMultiJSON(filePath *C.cchar_t) int {
	err := log.SetMultiJSON(C.GoString(filePath))
	if err != nil {
		return -1
	}
	return 0
}
