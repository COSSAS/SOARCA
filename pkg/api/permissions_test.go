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

func TestOnlyPingAndSwaggerArePublic(t *testing.T) {
	for key, rule := range Permissions {
		if rule.Public && key != "GET /status/ping" && key != "GET /swagger/*any" {
			t.Errorf("%s is public", key)
		}
		if !rule.Public && rule.Permission == "" {
			t.Errorf("%s needs a permission or Public", key)
		}
	}
}
