package auth

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Authenticator) redirectToOIDCLogin() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := state()
		if err != nil {
			// JSONErrorStatus(w, http.StatusInternalServerError, errors.New("generating state failed"))
			// JSONErrorStatus(w, http.StatusInternalServerError, err)]w.WriteHeader(status)
			b, _ := json.Marshal(map[string]string{"error": err.Error()})
			_, _ = w.Write(b)
			return
		}

		// a.jar.setStateSession(r, w, state)

		http.Redirect(w, r, a.config.OAuth2.AuthCodeURL(state), http.StatusFound)
	}
}

func (a *Authenticator) authConfig() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		b, _ := json.Marshal(map[string]any{
			// "simple": a.config.SimpleAuthEnable,
			"oidc": "true",
		})

		_, _ = writer.Write(b)
	}
}

func (a *Authenticator) AuthRoutes(route *gin.Engine) {
	authRoutes := route.Group("/auth")
	{
		authRoutes.GET("/oidclogin", gin.WrapF(a.redirectToOIDCLogin()))
	}
	// server.GET("/config", a.authConfig())
	// server.GET("/callback", a.Callback())

	// server.Post("/logout", a.logout())
}
