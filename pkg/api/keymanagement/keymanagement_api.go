package keymanagement

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/capability/ssh"
	"soarca/pkg/models/api"
	"strconv"

	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

type KeyManagementHandler struct {
	Manager *ssh.KeyManagement
}

func NewKeyManagementHandler(manager *ssh.KeyManagement) *KeyManagementHandler {
	return &KeyManagementHandler{
		Manager: manager,
	}
}

func (handler *KeyManagementHandler) GetKeys(context *gin.Context) {
	keyInfo := handler.Manager.ListAllNames()
	context.JSON(http.StatusOK, keyInfo)
}

func (handler *KeyManagementHandler) Refresh(context *gin.Context) {
	err := handler.Manager.Refresh()
	if err != nil {
		log.Trace("Refreshing keys has failed: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to refresh keys", "POST /keymanagement/refresh")
		return
	}
	context.JSON(http.StatusOK, Empty{})
}

func (handler *KeyManagementHandler) AddKey(context *gin.Context) {
	keyname := context.Param("keyname")
	jsonData, err := io.ReadAll(context.Request.Body)
	if err != nil {
		log.Trace("Submit key has failed: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to read json on server side", "POST /keymanagement")
		return
	}
	var key api.KeyManagementKey
	if err := json.Unmarshal(jsonData, &key); err != nil {
		log.Trace("Submit key failed to unmarshal: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to marshall json on server side", "POST /keymanagement")
		return
	}
	if err := handler.Manager.Insert(key.Public, key.Private, keyname); err != nil {
		log.Trace("Submit key failed to insert: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to insert key on server side", "POST /keymanagement")
		return

	}
	context.JSON(http.StatusOK, Empty{})
}

func (handler *KeyManagementHandler) RevokeKey(context *gin.Context) {
	keyname := context.Param("keyname")
	if err := handler.Manager.Revoke(keyname); err != nil {
		log.Trace("Revoke key failed: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to revoke key", "DELETE /keymanagement")
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": fmt.Sprintf("Successfuly removed key %s from SOARCA listing. To permanently remove key, delete the revoked key files in key management directory", keyname),
	})
}

func SendErrorResponse(context *gin.Context, status int, message string, orginal_call string) {
	msg := gin.H{
		"status":        strconv.Itoa(status),
		"message":       message,
		"original-call": orginal_call,
	}
	context.JSON(status, msg)
}
