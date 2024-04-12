package authentik_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	auth "soarca/utils/auth"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type OIDCProvider interface {
	AddClient(id, redirectURL string) string
	Close()
	URL() string
}

func newAuthTestServerPort(t *testing.T, config *auth.OidsConfig, oidcProvider OIDCProvider, host, port string) *httptest.Server {
	t.Helper()

	// init test server
	testServer := httptest.NewUnstartedServer(nil)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		t.Fatal(err)
	}
	testServer.Listener = listener

	redirectURL := fmt.Sprintf("http://%s:%s/auth/callback", host, port)
	secret := oidcProvider.AddClient(fmt.Sprintf("apitest-%s", port), redirectURL)

	config.OIDCIssuer = oidcProvider.URL()
	config.OAuth2.ClientID = fmt.Sprintf("apitest-%s", port)
	config.OAuth2.ClientSecret = secret
	config.OAuth2.RedirectURL = redirectURL

	testServer.Config.Handler = newAuthServer(t, config)

	return testServer
}

func newAuthServer(t *testing.T, config *auth.OidsConfig) *gin.Engine {
	t.Helper()

	ctx := context.Background()

	authenticator, err := auth.NewAuthenticatator(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	server := gin.New()
	server.Use(
	//	authenticator.Authenticate(),
	// authenticator.AuthorizeBlockedUser(),
	// authenticator.AuthorizePermission("automation:read"),
	).
		GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

	authenticator.AuthRoutes(server)

	return server
}

type HTTPResponse struct {
	StatusCode int
	Body       string
	BodyRegexp string
}

func TestOIDC(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	t.Parallel()

	setup := []struct {
		name      string
		oidcSetup func() OIDCProvider
		authSetup func(*testing.T, *auth.OidsConfig, OIDCProvider, string, string) *httptest.Server
		host      string
		portLow   int
		login     func(client *http.Client, port string, initialResp *http.Response) (*http.Response, error)
	}{
		{name: "authentic", oidcSetup: newAuthentik, authSetup: newAuthTestServerPort, host: "localhost", portLow: 92, login: authentikLogin},
	}

	tests := []struct {
		name     string
		portHigh int
		config   *auth.OidsConfig
		want     *HTTPResponse
	}{
		{
			name:     "success",
			portHigh: 90,
			config: &auth.OidsConfig{
				OAuth2: &oauth2.Config{
					Scopes: []string{oidc.ScopeOpenID, "profile", "email"}, // TODO: add groups
				},
				UserConfig: &auth.UserConfig{
					OIDCClaimUsername: "preferred_username",
					OIDCClaimEmail:    "email",
					OIDCClaimName:     "name",
					OIDCClaimGroups:   "groups",
					AuthDefaultRoles:  []string{"analyst"},
				},
			},
			want: &HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       "success",
			},
		},
	}
	for _, su := range setup {
		setup := su
		t.Run(su.name, func(t *testing.T) {
			t.Parallel()

			// create oidc test server
			oidcServer := setup.oidcSetup()
			defer oidcServer.Close()

			// for _, tt := range tests {
			// 	tt := tt
			// 	t.Run(tt.name, func(t *testing.T) {
			// 		// create test server
			// 		port := fmt.Sprint(tt.portHigh*100 + su.portLow)
			// 		// authServer := su.authSetup(t, tt.config.Clone(), oidcServer, su.host, port)
			// 		// authServer.Start()
			// 		// defer authServer.Close()

			// 		// create cookie jar
			// 		// client, err := testClient(t)
			// 		if err != nil {
			// 			t.Error(err)
			// 		}

			// 		// perform initial request
			// 		initialResp, err := client.Get(authServer.URL + "/")
			// 		if err != nil {
			// 			t.Fatal(err)
			// 		}
			// 		defer initialResp.Body.Close()

			// 		// send password
			// 		loginResp, err := su.login(client, port, initialResp)
			// 		if err != nil {
			// 			t.Fatal(err)
			// 		}
			// 		defer loginResp.Body.Close()

			// 		assertResult(t, loginResp, tt.want)
			// 	})
			// }
		})
	}
}
