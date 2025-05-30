package rest

import (
	"errors"
	"github.com/gin-gonic/gin"
	pwd "github.com/image357/password"
	"github.com/image357/password/log"
	"net/http"
	pathlib "path"
)

type multiOverwriteData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	Id          string `form:"id" json:"id" xml:"id"  binding:"required"`
	Password    string `form:"password" json:"password" xml:"password"  binding:"required"`
}

type multiGetData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	Id          string `form:"id" json:"id" xml:"id"  binding:"required"`
}

type multiCheckData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	Id          string `form:"id" json:"id" xml:"id"  binding:"required"`
	Password    string `form:"password" json:"password" xml:"password"  binding:"required"`
}

type multiSetData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	Id          string `form:"id" json:"id" xml:"id"  binding:"required"`
	OldPassword string `form:"oldPassword" json:"oldPassword" xml:"oldPassword"  binding:"required"`
	NewPassword string `form:"newPassword" json:"newPassword" xml:"newPassword"  binding:"required"`
}

type multiUnsetData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	Id          string `form:"id" json:"id" xml:"id"  binding:"required"`
	Password    string `form:"password" json:"password" xml:"password"  binding:"required"`
}

type multiExistsData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	Id          string `form:"id" json:"id" xml:"id"  binding:"required"`
}

type multiListData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
}

type multiDeleteData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
	Id          string `form:"id" json:"id" xml:"id"  binding:"required"`
}

type multiCleanData struct {
	AccessToken string `form:"accessToken" json:"accessToken" xml:"accessToken"  binding:"required"`
}

// StartMultiService creates a multi password REST service.
// The service binds to
// "/prefix/overwrite" (PUT),
// "/prefix/get" (GET),
// "/prefix/check" (GET),
// "/prefix/set" (PUT),
// "/prefix/unset" (DELETE),
// "/prefix/exists" (GET),
// "/prefix/list" (GET),
// "/prefix/delete" (DELETE),
// "/prefix/clean" (DELETE).
// The callback of type TestAccessFunc will be called for every request to determine access.
func StartMultiService(bindAddress string, prefix string, key string, callback TestAccessFunc) error {
	// prepare arguments
	prefix = preparePrefix(prefix)

	// setup service
	engine, service, err := setupService(bindAddress, prefix, key, callback)
	if err != nil {
		return err
	}

	// inject current default manager into callbacks
	manager := pwd.GetDefaultManager()
	localOverwriteCallback := func(c *gin.Context) { multiOverwriteCallback(c, manager, service) }
	localGetCallback := func(c *gin.Context) { multiGetCallback(c, manager, service) }
	localCheckCallback := func(c *gin.Context) { multiCheckCallback(c, manager, service) }
	localSetCallback := func(c *gin.Context) { multiSetCallback(c, manager, service) }
	localUnsetCallback := func(c *gin.Context) { multiUnsetCallback(c, manager, service) }
	localExistsCallback := func(c *gin.Context) { multiExistsCallback(c, manager, service) }
	localListCallback := func(c *gin.Context) { multiListCallback(c, manager, service) }
	localDeleteCallback := func(c *gin.Context) { multiDeleteCallback(c, manager, service) }
	localCleanCallback := func(c *gin.Context) { multiCleanCallback(c, manager, service) }

	// setup REST endpoints
	engine.PUT(pathlib.Join("/", prefix, "/overwrite"), localOverwriteCallback)
	engine.GET(pathlib.Join("/", prefix, "/get"), localGetCallback)
	engine.GET(pathlib.Join("/", prefix, "/check"), localCheckCallback)
	engine.PUT(pathlib.Join("/", prefix, "/set"), localSetCallback)
	engine.DELETE(pathlib.Join("/", prefix, "/unset"), localUnsetCallback)
	engine.GET(pathlib.Join("/", prefix, "/exists"), localExistsCallback)
	engine.GET(pathlib.Join("/", prefix, "/list"), localListCallback)
	engine.DELETE(pathlib.Join("/", prefix, "/delete"), localDeleteCallback)
	engine.DELETE(pathlib.Join("/", prefix, "/clean"), localCleanCallback)

	go func() {
		log.Info(
			restStartedLogMsg,
			"addr", bindAddress,
			"prefix", prefix,
			"type", "multi",
			"TLS", useTLS,
		)

		var err error = nil
		if useTLS {
			err = service.server.ListenAndServeTLS(certFileTLS, keyFileTLS)

		} else {
			err = service.server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(
				restStoppedLogMsg,
				"error", err,
				"addr", bindAddress,
				"prefix", prefix,
				"type", "multi",
				"TLS", useTLS,
			)
		}
		delete(services, service.name)
	}()
	return nil
}

func multiOverwriteCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data multiOverwriteData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(data.Id)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Overwrite(data.Id, data.Password, getStorageKey(s))
	if err != nil {
		log.Error("rest: Overwrite failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func multiGetCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data multiGetData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(data.Id)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	password, err := m.Get(data.Id, getStorageKey(s))
	if err != nil {
		log.Error("rest: Get failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"password": password})
}

func multiCheckCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data multiCheckData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(data.Id)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	result, err := m.Check(data.Id, data.Password, getStorageKey(s))
	if err != nil {
		log.Error("rest: Check failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func multiSetCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data multiSetData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(data.Id)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Set(data.Id, data.OldPassword, data.NewPassword, getStorageKey(s))
	if err != nil {
		log.Error("rest: Set failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func multiUnsetCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data multiUnsetData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(data.Id)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Unset(data.Id, data.Password, getStorageKey(s))
	if err != nil {
		log.Error("rest: Unset failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func multiExistsCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data multiExistsData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(data.Id)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	result, err := m.Exists(data.Id)
	if err != nil {
		log.Error("rest: Exists failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func multiListCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data multiListData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId("")
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	list, err := m.List()
	if err != nil {
		log.Error("rest: List failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ids": list})
}

func multiDeleteCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data multiDeleteData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId(data.Id)
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Delete(data.Id)
	if err != nil {
		log.Error("rest: Delete failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func multiCleanCallback(c *gin.Context, m *pwd.Manager, s *restService) {
	logContext(c)

	var data multiCleanData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(processDataLogMsg, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	ip := c.ClientIP()
	url := c.Request.URL.String()
	id := pwd.NormalizeId("")
	if !s.hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = m.Clean()
	if err != nil {
		log.Error("rest: Clean failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
