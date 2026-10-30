package auth

import (
	"net/http"

	"soarca/pkg/models/api"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	config api.AuthConfig
}

func NewAuthHandler(config api.AuthConfig) *AuthHandler {
	return &AuthHandler{config: config}
}

// GetConfig returns the providers users can sign in with.
//
//	@Summary	gets the sign-in options
//	@Tags		auth
//	@Produce	json
//	@success	200	{object}	api.AuthConfig
//	@Router		/auth/config [GET]
func (handler *AuthHandler) GetConfig(ginContext *gin.Context) {
	ginContext.JSON(http.StatusOK, handler.config)
}

// GetMe returns the caller.
//
//	@Summary	gets the signed-in caller
//	@Tags		auth
//	@Produce	json
//	@success	200	{object}	api.AuthUser
//	@failure	401	{object}	api.Error
//	@Security	BearerAuth
//	@Router		/auth/me [GET]
func (handler *AuthHandler) GetMe(ginContext *gin.Context) {
	ginContext.Header("Cache-Control", "no-store")
	ginContext.JSON(http.StatusOK, api.AuthUser{Principal: PrincipalFromContext(ginContext)})
}
