package auth

import (
	"net/http"

	"soarca/pkg/models/api"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	issuers     []string
	scopes      []string
	resourceUrl string
}

// NewAuthHandler publishes issuers and scopes as the protected resource
// metadata. resourceUrl is SOARCA's public URL; empty takes it from each
// request.
func NewAuthHandler(issuers []string, scopes []string, resourceUrl string) *AuthHandler {
	if issuers == nil {
		issuers = []string{}
	}
	return &AuthHandler{issuers: issuers, scopes: scopes, resourceUrl: resourceUrl}
}

// GetProtectedResourceMetadata returns the OAuth protected resource metadata
// (RFC 9728): the identity providers SOARCA accepts tokens from and the
// scopes to request. An empty authorization_servers list means
// authentication is disabled.
//
//	@Summary	gets the OAuth protected resource metadata
//	@Tags		auth
//	@Produce	json
//	@success	200	{object}	api.ProtectedResourceMetadata
//	@Router		/.well-known/oauth-protected-resource [GET]
func (handler *AuthHandler) GetProtectedResourceMetadata(ginContext *gin.Context) {
	ginContext.JSON(http.StatusOK, api.ProtectedResourceMetadata{
		Resource:               ResourceUrl(ginContext, handler.resourceUrl),
		ResourceName:           "SOARCA",
		AuthorizationServers:   handler.issuers,
		ScopesSupported:        handler.scopes,
		BearerMethodsSupported: []string{"header"},
	})
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
