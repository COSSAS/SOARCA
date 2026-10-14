package auth

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"soarca/internal/logger"
	apiError "soarca/pkg/api/error"
	"soarca/pkg/auth"
	"soarca/pkg/auth/roles"

	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

const principalKey = "soarca.auth.principal"

// Middleware authenticates each request's bearer credential and stores the
// caller for PrincipalFromContext. It responds 401 when that fails.
func Middleware(authenticator auth.IAuthenticator) gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		credential := bearerCredential(ginContext.GetHeader("Authorization"))
		principal, err := authenticator.Authenticate(ginContext.Request.Context(), credential)
		if err != nil {
			log.Debug(fmt.Sprintf("authentication failed for %s: %s", ginContext.FullPath(), err))
			ginContext.Header("WWW-Authenticate", `Bearer realm="soarca"`)
			apiError.SendErrorResponse(ginContext, http.StatusUnauthorized, "unauthenticated", originalCall(ginContext), "")
			ginContext.Abort()
			return
		}
		ginContext.Set(principalKey, principal)
		ginContext.Next()
	}
}

// RequireAll lets a request through only if the caller holds every permission.
func RequireAll(permissions ...roles.Permission) gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		principal := PrincipalFromContext(ginContext)
		for _, permission := range permissions {
			if !principal.Has(permission) {
				deny(ginContext, principal, string(permission))
				return
			}
		}
		ginContext.Next()
	}
}

// RequireAny lets a request through if the caller holds at least one permission.
func RequireAny(permissions ...roles.Permission) gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		principal := PrincipalFromContext(ginContext)
		names := make([]string, 0, len(permissions))
		for _, permission := range permissions {
			if principal.Has(permission) {
				ginContext.Next()
				return
			}
			names = append(names, string(permission))
		}
		deny(ginContext, principal, strings.Join(names, " or "))
	}
}

// PrincipalFromContext returns the caller, or nil when Middleware didn't run.
func PrincipalFromContext(ginContext *gin.Context) *auth.Principal {
	value, _ := ginContext.Get(principalKey)
	principal, _ := value.(*auth.Principal)
	return principal
}

func deny(ginContext *gin.Context, principal *auth.Principal, required string) {
	if principal == nil {
		apiError.SendErrorResponse(ginContext, http.StatusUnauthorized, "unauthenticated", originalCall(ginContext), "")
		ginContext.Abort()
		return
	}
	log.Info(fmt.Sprintf("permission denied: %s (client %q) needs %s for %s",
		principal.Self(), principal.ClientId, required, ginContext.FullPath()))
	apiError.SendErrorResponse(ginContext, http.StatusForbidden, "forbidden: requires "+required, originalCall(ginContext), "")
	ginContext.Abort()
}

func originalCall(ginContext *gin.Context) string {
	return ginContext.Request.Method + " " + ginContext.Request.URL.Path
}

func bearerCredential(header string) string {
	const scheme = "bearer "
	if len(header) > len(scheme) && strings.EqualFold(header[:len(scheme)], scheme) {
		return strings.TrimSpace(header[len(scheme):])
	}
	return ""
}
