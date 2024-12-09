package rest

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	pwd "github.com/image357/password"
	"github.com/image357/password/log"
	"net/http"
	pathlib "path"
	"time"
)

const defaultId = "default"

const accessDeniedLogMsg = "rest: access denied"
const processDataLogMsg = "rest: cannot process data"
const restStartedLogMsg = "rest: service started"
const restStoppedLogMsg = "rest: service stopped"

var restRunningErr = errors.New("rest service already running")
var restStoppedErr = errors.New("rest service already stopped")

// TestAccessFunc is a callback signature.
// The callback will be called by the rest service for every request to determine access based on the accessToken.
type TestAccessFunc func(token string, ip string, resource string, id string) bool

// restService contains all necessary information for external handling of a REST service.
type restService struct {
	name             string
	server           *http.Server
	storageKeyBytes  []byte
	storageKeySecret []byte
	hasAccess        TestAccessFunc
}

// services contains the global map of all started REST servers.
var services = make(map[string]*restService)

// getStorageKey decrypts the storage key that was set by StartSimpleService or StartMultiService.
func getStorageKey(service *restService) string {
	return pwd.DecryptOTP(service.storageKeyBytes, service.storageKeySecret)
}

type simpleOverwriteData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	Password    string `form:"password" json:"password" xml:"password"  binding:"required"`
}

type simpleGetData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
}

type simpleCheckData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	Password    string `form:"password" json:"password" xml:"password"  binding:"required"`
}

type simpleSetData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	OldPassword string `form:"oldPassword" json:"oldPassword" xml:"oldPassword"  binding:"required"`
	NewPassword string `form:"newPassword" json:"newPassword" xml:"newPassword"  binding:"required"`
}

type simpleUnsetData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	Password    string `form:"password" json:"password" xml:"password"  binding:"required"`
}

type simpleExistsData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
}

type simpleDeleteData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
}

// setupService returns a basic gin.Engine without any endpoint configuration.
func setupService(bindAddress string, prefix string, key string, callback TestAccessFunc) (*gin.Engine, *restService, error) {
	name := pathlib.Clean(bindAddress + "/" + prefix)
	_, ok := services[name]
	if ok {
		return nil, nil, restRunningErr
	}

	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(gin.Recovery())

	s := new(restService)
	s.name = name
	s.server = &http.Server{
		Addr:    bindAddress,
		Handler: e,
	}
	s.storageKeyBytes, s.storageKeySecret = pwd.EncryptOTP(key)
	s.hasAccess = callback

	services[name] = s
	return e, s, nil
}

// logContext will be called by the rest service for every request.
func logContext(c *gin.Context) {
	log.Debug(
		"rest: request",
		"ip", c.ClientIP(),
		"resource", c.Request.URL.String(),
	)
}

// StartSimpleService creates a single password rest service.
// The service binds to
// "/prefix/overwrite" (PUT),
// "/prefix/get" (GET),
// "/prefix/check" (GET),
// "/prefix/set" (PUT),
// "/prefix/unset" (DELETE),
// "/prefix/exists" (GET),
// "/prefix/delete" (DELETE).
// The callback of type TestAccessFunc will be called for every request to determine access.
func StartSimpleService(bindAddress string, prefix string, key string, callback TestAccessFunc) error {
	engine, service, err := setupService(bindAddress, prefix, key, callback)
	if err != nil {
		return err
	}

	// inject current default manager and service into callbacks
	manager := pwd.GetDefaultManager()
	localOverwriteCallback := func(c *gin.Context) { simpleOverwriteCallback(c, manager, service) }
	localGetCallback := func(c *gin.Context) { simpleGetCallback(c, manager, service) }
	localCheckCallback := func(c *gin.Context) { simpleCheckCallback(c, manager, service) }
	localSetCallback := func(c *gin.Context) { simpleSetCallback(c, manager, service) }
	localUnsetCallback := func(c *gin.Context) { simpleUnsetCallback(c, manager, service) }
	localExistsCallback := func(c *gin.Context) { simpleExistsCallback(c, manager, service) }
	localDeleteCallback := func(c *gin.Context) { simpleDeleteCallback(c, manager, service) }

	// setup rest endpoints
	engine.PUT(pathlib.Join("/", pwd.NormalizeId(prefix), "/overwrite"), localOverwriteCallback)
	engine.GET(pathlib.Join("/", pwd.NormalizeId(prefix), "/get"), localGetCallback)
	engine.GET(pathlib.Join("/", pwd.NormalizeId(prefix), "/check"), localCheckCallback)
	engine.PUT(pathlib.Join("/", pwd.NormalizeId(prefix), "/set"), localSetCallback)
	engine.DELETE(pathlib.Join("/", pwd.NormalizeId(prefix), "/unset"), localUnsetCallback)
	engine.GET(pathlib.Join("/", pwd.NormalizeId(prefix), "/exists"), localExistsCallback)
	engine.DELETE(pathlib.Join("/", pwd.NormalizeId(prefix), "/delete"), localDeleteCallback)

	go func() {
		log.Info(restStartedLogMsg, "addr", bindAddress, "prefix", prefix, "type", "simple")
		err := service.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(restStoppedLogMsg, "error", err)
		}
		delete(services, service.name)
	}()
	return nil
}

// StopService will block execution and try to gracefully shut down any rest service during the timeout period.
// The service is guaranteed to be closed at the end of the timeout.
func StopService(timeout int, bindAddress string, prefix string) error {
	name := pathlib.Clean(bindAddress + "/" + prefix)
	service, ok := services[name]
	if !ok {
		delete(services, name)
		return restStoppedErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	defer cancel()

	err := service.server.Shutdown(ctx)
	if err != nil {
		log.Warn(restStoppedLogMsg, "error", err)
	}
	err = service.server.Close()
	if err != nil {
		log.Warn(restStoppedLogMsg, "error", err)
	}

	// cleanup service
	service.server = nil
	service.storageKeyBytes, service.storageKeySecret = nil, nil
	service.hasAccess = nil
	delete(services, name)

	log.Info(restStoppedLogMsg)
	return nil
}

func simpleOverwriteCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data simpleOverwriteData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(defaultId)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Overwrite(defaultId, data.Password, getStorageKey(s))
	if err != nil {
		log.Error("rest: Overwrite failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func simpleGetCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data simpleGetData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(defaultId)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	password, err := m.Get(defaultId, getStorageKey(s))
	if err != nil {
		log.Error("rest: Get failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"password": password})
}

func simpleCheckCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data simpleCheckData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(defaultId)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	result, err := m.Check(defaultId, data.Password, getStorageKey(s))
	if err != nil {
		log.Error("rest: Check failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func simpleSetCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data simpleSetData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(defaultId)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Set(defaultId, data.OldPassword, data.NewPassword, getStorageKey(s))
	if err != nil {
		log.Error("rest: Set failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func simpleUnsetCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data simpleUnsetData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(defaultId)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Unset(defaultId, data.Password, getStorageKey(s))
	if err != nil {
		log.Error("rest: Unset failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func simpleExistsCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data simpleExistsData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(defaultId)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	result, err := m.Exists(defaultId)
	if err != nil {
		log.Error("rest: Exists failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func simpleDeleteCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data simpleDeleteData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(defaultId)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Delete(defaultId)
	if err != nil {
		log.Error("rest: Delete failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
