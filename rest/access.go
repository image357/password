package rest

import (
	"github.com/image357/password/log"
)

var debugAccessSuccess = true

// DebugAccessCallback returns the global variable debugAccessSuccess.
// It will log the arguments to the package logger in debug level.
func DebugAccessCallback(token string, ip string, resource string, id string) bool {
	log.Debug("rest: access callback", "ip", ip, "resource", resource, "id", id, "token", token)
	return debugAccessSuccess
}

// FullAccessCallback will grant access to every rest request.
func FullAccessCallback(_ string, _ string, _ string, _ string) bool {
	return true
}
