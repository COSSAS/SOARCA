package keymanagement_api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/keymanagement"
	"soarca/pkg/models/api"
	"strconv"

	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Trace, "", logger.Json)
}

type KeyManagementHandler struct {
	Manager *keymanagement.KeyManagement
}

func NewKeyManagementHandler(manager *keymanagement.KeyManagement) *KeyManagementHandler {
	return &KeyManagementHandler{
		Manager: manager,
	}
}

// GetKeys GET handler for obtaining all keys in the key managements system
//
//	@Summary	gets all keys from the KMS
//	@Schemes
//	@Description	return all keys in the KMS
//	@Tags			keymanagement
//	@Produce		json
//	@success		200	{array}		string
//	@failure		400	{object}	api.Error
//	@Router			/keymanagement/ [GET]
func (handler *KeyManagementHandler) GetKeys(context *gin.Context) {
	keyInfo := handler.Manager.ListAllNames()
	log.Trace("Listing all key names")
	context.JSON(http.StatusOK, keyInfo)
}

// AddKey PUT handler to add key to KMS
//
//	@Summary add key to KMS
//	@Schemes
//	@Description	adds a key to the KMS; load key into cache and write to database
//	@Tags			keymanagement
//	@Param			data	body		api.KeyManagementKey	true	"key"
//	@Produce		json
//	@success		200	{json}		Empty
//	@failure		400	{object}	api.Error
//	@Router			/keymanagement/:keyname/ [PUT]
func (handler *KeyManagementHandler) AddKey(context *gin.Context) {
	keyname := context.Param("keyname")
	jsonData, err := io.ReadAll(context.Request.Body)
	if err != nil {
		log.Trace("Submit key has failed: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to read json on server side", "PUT /keymanagement/:keyname")
		return
	}
	var key api.KeyManagementKey
	if err := json.Unmarshal(jsonData, &key); err != nil {
		log.Trace("Submit key failed to unmarshal: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to marshall json on server side", "PUT /keymanagement/:keyname")
		return
	}
	if err := handler.Manager.Insert(key.Public, key.Private, key.Passphrase, keyname); err != nil {
		log.Trace("Submit key failed to insert: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to insert key on server side", "PUT /keymanagement/:keyname")
		return

	}
	log.Trace("Inserted key ", keyname)
	context.JSON(http.StatusOK, Empty{})
}

// UpdateKey PATCH handler to update key in KMS
//
//	@Summary update key in KMS
//	@Schemes
//	@Description	update a key in the KMS; load key into cache and write to database
//	@Tags			keymanagement
//	@Param			data	body		api.KeyManagementKey	true	"key"
//	@Produce		json
//	@success		200	{json}		Empty
//	@failure		400	{object}	api.Error
//	@Router			/keymanagement/:keyname/ [PATCH]
func (handler *KeyManagementHandler) UpdateKey(context *gin.Context) {
	keyname := context.Param("keyname")
	jsonData, err := io.ReadAll(context.Request.Body)
	if err != nil {
		log.Trace("Update key has failed: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to read json on server side", "PATCH /keymanagement/:keyname")
		return
	}
	var key api.KeyManagementKey
	if err := json.Unmarshal(jsonData, &key); err != nil {
		log.Trace("Update key failed to unmarshal: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to marshall json on server side", "PATCH /keymanagement/:keyname")
		return
	}
	if err := handler.Manager.Update(key.Public, key.Private, key.Passphrase, keyname); err != nil {
		log.Trace("Update key failed to insert: ", err.Error())
		SendErrorResponse(context, http.StatusBadRequest, "Failed to update key on server side", "PATCH /keymanagement/:keyname")
		return

	}
	log.Trace("Updated key ", keyname)
	context.JSON(http.StatusOK, Empty{})
}

// RevokeKey DELETE handler to remove key from KMS
//
//	@Summary remove key from KMS
//	@Schemes
//	@Description	revokes the key by moving it to .revoked and renaming it
//	@Tags			keymanagement
//	@Produce		json
//	@success		200	{json}		Empty
//	@failure		400	{object}	api.Error
//	@Router			/keymanagement/:keyname/ [DELETE]
func (handler *KeyManagementHandler) RevokeKey(context *gin.Context) {
	keyname := context.Param("keyname")
	handler.Manager.Revoke(keyname)
	context.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": fmt.Sprintf("Removed key %s from SOARCA listing", keyname),
	})
	log.Trace("Removed key ", keyname)
}

func SendErrorResponse(context *gin.Context, status int, message string, orginal_call string) {
	msg := gin.H{
		"status":        strconv.Itoa(status),
		"message":       message,
		"original-call": orginal_call,
	}
	context.JSON(status, msg)
}
