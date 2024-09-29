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

// StartMultiService creates a multi password rest service.
// The service binds to "/prefix/overwrite" (PUT), "/prefix/get" (GET), "/prefix/check" (GET), "/prefix/set" (PUT), "/prefix/unset" (DELETE), "/prefix/list" (GET), "/prefix/delete" (DELETE), "/prefix/clean" (DELETE).
// The callback of type TestAccessFunc will be called for every request to determine access.
func StartMultiService(bindAddress string, prefix string, key string, callback TestAccessFunc) error {
	engine, err := setupEngine(bindAddress, key, callback)
	if err != nil {
		return err
	}

	engine.PUT(pathlib.Join("/", pwd.NormalizeId(prefix), "/overwrite"), multiOverwriteCallback)
	engine.GET(pathlib.Join("/", pwd.NormalizeId(prefix), "/get"), multiGetCallback)
	engine.GET(pathlib.Join("/", pwd.NormalizeId(prefix), "/check"), multiCheckCallback)
	engine.PUT(pathlib.Join("/", pwd.NormalizeId(prefix), "/set"), multiSetCallback)
	engine.DELETE(pathlib.Join("/", pwd.NormalizeId(prefix), "/unset"), multiUnsetCallback)
	engine.GET(pathlib.Join("/", pwd.NormalizeId(prefix), "/list"), multiListCallback)
	engine.DELETE(pathlib.Join("/", pwd.NormalizeId(prefix), "/delete"), multiDeleteCallback)
	engine.DELETE(pathlib.Join("/", pwd.NormalizeId(prefix), "/clean"), multiCleanCallback)

	go func() {
		log.Info(restStartedLogMsg, "addr", bindAddress, "prefix", prefix, "type", "multi")
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Warn(restStoppedLogMsg, "error", err)
		}
	}()
	return nil
}

func multiOverwriteCallback(c *gin.Context) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = pwd.Overwrite(data.Id, data.Password, storageKey)
	if err != nil {
		log.Error("rest: Overwrite failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func multiGetCallback(c *gin.Context) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	password, err := pwd.Get(data.Id, storageKey)
	if err != nil {
		log.Error("rest: Get failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"password": password})
}

func multiCheckCallback(c *gin.Context) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	result, err := pwd.Check(data.Id, data.Password, storageKey)
	if err != nil {
		log.Error("rest: Check failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func multiSetCallback(c *gin.Context) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = pwd.Set(data.Id, data.OldPassword, data.NewPassword, storageKey)
	if err != nil {
		log.Error("rest: Set failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func multiUnsetCallback(c *gin.Context) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = pwd.Unset(data.Id, data.Password, storageKey)
	if err != nil {
		log.Error("rest: Unset failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func multiListCallback(c *gin.Context) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	list, err := pwd.List()
	if err != nil {
		log.Error("rest: List failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ids": list})
}

func multiDeleteCallback(c *gin.Context) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = pwd.Delete(data.Id)
	if err != nil {
		log.Error("rest: Delete failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func multiCleanCallback(c *gin.Context) {
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
	if !hasAccess(data.AccessToken, ip, url, id) {
		log.Warn(accessDeniedLogMsg, "ip", ip, "resource", url, "id", id, "token", data.AccessToken)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	err = pwd.Clean()
	if err != nil {
		log.Error("rest: Clean failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
