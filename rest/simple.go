package rest

import (
	"context"
	"errors"
	"fmt"
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
const restRunningErrMsg = "rest service already running"

var server *http.Server
var storageKeyBytes []byte
var storageKeySecret []byte

// getStorageKey returns the storage key that was set StartSimpleService or StartMultiService.
func getStorageKey() string {
	return pwd.DecryptOTP(storageKeyBytes, storageKeySecret)
}

// TestAccessFunc is a callback signature.
// The callback will be called by the rest service for every request to determine access based on the accessToken.
type TestAccessFunc func(token string, ip string, resource string, id string) bool

var hasAccess TestAccessFunc

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

// setupEngine returns a basic gin.Engine without any endpoint configuration.
func setupEngine(bindAddress string, key string, callback TestAccessFunc) (*gin.Engine, error) {
	if server != nil {
		return nil, fmt.Errorf(restRunningErrMsg)
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	server = &http.Server{
		Addr:    bindAddress,
		Handler: engine,
	}

	storageKeyBytes, storageKeySecret = pwd.EncryptOTP(key)
	hasAccess = callback

	return engine, nil
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
// The service binds to "/prefix/overwrite" (PUT), "/prefix/get" (GET), "/prefix/check" (GET), "/prefix/set" (PUT), "/prefix/unset" (DELETE), "/prefix/delete" (DELETE).
// The callback of type TestAccessFunc will be called for every request to determine access.
// Warning: calling this function will reset the default password manager and register the current one as "rest manger: bindAddress/prefix".
func StartSimpleService(bindAddress string, prefix string, key string, callback TestAccessFunc) error {
	engine, err := setupEngine(bindAddress, key, callback)
	if err != nil {
		return err
	}

	// inject current default manager into callbacks
	m := pwd.GetDefaultManager()
	localOverwriteCallback := func(c *gin.Context) { simpleOverwriteCallback(c, m) }
	localGetCallback := func(c *gin.Context) { simpleGetCallback(c, m) }
	localCheckCallback := func(c *gin.Context) { simpleCheckCallback(c, m) }
	localSetCallback := func(c *gin.Context) { simpleSetCallback(c, m) }
	localUnsetCallback := func(c *gin.Context) { simpleUnsetCallback(c, m) }
	localExistsCallback := func(c *gin.Context) { simpleExistsCallback(c, m) }
	localDeleteCallback := func(c *gin.Context) { simpleDeleteCallback(c, m) }
	pwd.RegisterDefaultManager("rest manager: " + pathlib.Join(bindAddress+"/"+prefix))

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
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Warn(restStoppedLogMsg, "error", err)
		}
	}()
	return nil
}

// StopService will block execution and try to gracefully shut down any rest service during the timeout period.
// The service is guaranteed to be closed at the end of the timeout.
func StopService(timeout int) error {
	if server == nil {
		return fmt.Errorf("rest service already stopped")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		log.Warn(restStoppedLogMsg, "error", err)
	}
	err = server.Close()
	if err != nil {
		log.Warn(restStoppedLogMsg, "error", err)
	}

	server = nil
	storageKeyBytes, storageKeySecret = nil, nil
	hasAccess = nil

	log.Info(restStoppedLogMsg)
	return nil
}

func simpleOverwriteCallback(c *gin.Context, m *pwd.Manager) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Overwrite(defaultId, data.Password, getStorageKey())
	if err != nil {
		log.Error("rest: Overwrite failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func simpleGetCallback(c *gin.Context, m *pwd.Manager) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	password, err := m.Get(defaultId, getStorageKey())
	if err != nil {
		log.Error("rest: Get failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"password": password})
}

func simpleCheckCallback(c *gin.Context, m *pwd.Manager) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	result, err := m.Check(defaultId, data.Password, getStorageKey())
	if err != nil {
		log.Error("rest: Check failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func simpleSetCallback(c *gin.Context, m *pwd.Manager) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Set(defaultId, data.OldPassword, data.NewPassword, getStorageKey())
	if err != nil {
		log.Error("rest: Set failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func simpleUnsetCallback(c *gin.Context, m *pwd.Manager) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Unset(defaultId, data.Password, getStorageKey())
	if err != nil {
		log.Error("rest: Unset failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func simpleExistsCallback(c *gin.Context, m *pwd.Manager) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
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

func simpleDeleteCallback(c *gin.Context, m *pwd.Manager) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
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
