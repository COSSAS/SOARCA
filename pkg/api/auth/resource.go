package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// ProtectedResourcePath serves the RFC 9728 protected resource metadata.
const ProtectedResourcePath = "/.well-known/oauth-protected-resource"

// ResourceUrl returns SOARCA's public URL: configured when set, otherwise
// the scheme and host the request came in on.
func ResourceUrl(ginContext *gin.Context, configured string) string {
	if configured != "" {
		return configured
	}
	scheme := "http"
	if ginContext.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, ginContext.Request.Host)
}

// ResourceMetadataUrl returns where the protected resource metadata is
// served, for the WWW-Authenticate header of 401 responses.
func ResourceMetadataUrl(ginContext *gin.Context, configured string) string {
	return ResourceUrl(ginContext, configured) + ProtectedResourcePath
}
