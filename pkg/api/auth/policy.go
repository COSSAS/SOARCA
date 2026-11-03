package auth

import (
	"soarca/pkg/auth"
	"soarca/pkg/auth/roles"

	"github.com/gin-gonic/gin"
)

// Rule says who may call one route.
type Rule struct {
	// Public routes need no credential at all.
	Public bool
	// Permission the caller must hold. Ignored for public routes.
	Permission roles.Permission
}

// Policy maps routes to rules. Keys are the method and the route pattern as
// registered with gin, e.g. "GET /playbook/:id".
type Policy map[string]Rule

// RouteKey returns the Policy key for a method and a route pattern.
func RouteKey(method string, fullPath string) string {
	return method + " " + fullPath
}

// Protect authenticates and authorizes every request against policy. Install
// it with app.Use before registering routes, so it covers all of them:
//
//   - a public route passes without a credential;
//   - any other route needs a valid credential (401 otherwise) and the
//     permission its rule names (403 otherwise);
//   - a registered route missing from policy needs roles.Admin, so a new
//     route is closed until someone decides who may call it;
//   - a path that matches no route still needs a valid credential, so
//     unauthenticated callers can't probe which routes exist.
func Protect(authenticator auth.IAuthenticator, policy Policy) gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		fullPath := ginContext.FullPath()
		rule, known := policy[RouteKey(ginContext.Request.Method, fullPath)]
		if known && rule.Public {
			ginContext.Next()
			return
		}

		principal, ok := authenticate(ginContext, authenticator)
		if !ok {
			return
		}

		if fullPath != "" {
			required := roles.Admin
			if known {
				required = rule.Permission
			} else {
				log.Warning("no permission rule for ", RouteKey(ginContext.Request.Method, fullPath), ", requiring ", roles.Admin)
			}
			if !principal.Has(required) {
				deny(ginContext, principal, string(required))
				return
			}
		}
		ginContext.Next()
	}
}
