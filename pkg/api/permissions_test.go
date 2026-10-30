package api

import (
	"testing"

	apiAuth "soarca/pkg/api/auth"
	mock_database_controller "soarca/test/unittest/mocks/mock_controller/database"
	mock_playbook "soarca/test/unittest/mocks/mock_playbook_database"

	"github.com/gin-gonic/gin"
)

// allRoutes registers every route the way the controller does.
func allRoutes(t *testing.T) gin.RoutesInfo {
	gin.SetMode(gin.TestMode)
	app := gin.New()
	database := &mock_database_controller.Mock_Controller{}
	database.On("GetDatabaseInstance").Return(&mock_playbook.MockPlaybook{})

	if err := Api(app, nil, database); err != nil {
		t.Fatal(err)
	}
	if err := Database(app, database); err != nil {
		t.Fatal(err)
	}
	if err := Reporter(app, nil); err != nil {
		t.Fatal(err)
	}
	KeyManagement(app, nil)
	Manual(app, nil)
	Swagger(app)
	AuthRoutes(app, nil)
	return app.Routes()
}

func TestEveryRouteHasAPermission(t *testing.T) {
	for _, info := range allRoutes(t) {
		if _, ok := Permissions[apiAuth.RouteKey(info.Method, info.Path)]; !ok {
			t.Errorf("%s %s has no entry in Permissions", info.Method, info.Path)
		}
	}
}

func TestEveryPermissionHasARoute(t *testing.T) {
	registered := map[string]bool{}
	for _, info := range allRoutes(t) {
		registered[apiAuth.RouteKey(info.Method, info.Path)] = true
	}
	for key := range Permissions {
		if !registered[key] {
			t.Errorf("Permissions lists %s, which isn't a route", key)
		}
	}
}

func TestPublicRoutesAreExplicit(t *testing.T) {
	public := map[string]bool{
		"GET /status/ping":  true,
		"GET /swagger/*any": true,
		"GET /auth/config":  true,
	}
	for key, rule := range Permissions {
		if rule.Public != public[key] {
			t.Errorf("%s: Public is %v, want %v", key, rule.Public, public[key])
		}
		if !rule.Public && !rule.AuthenticatedOnly && rule.Permission == "" {
			t.Errorf("%s needs a permission, Public or AuthenticatedOnly", key)
		}
		if rule.AuthenticatedOnly && (rule.Public || rule.Permission != "") {
			t.Errorf("%s mixes AuthenticatedOnly with Public or a permission", key)
		}
	}
}
