package api

import (
	"net/http"

	apiAuth "soarca/pkg/api/auth"
	"soarca/pkg/auth/roles"
)

// Permissions says who may call each route. Keep it in step with the route
// table in docs/content/en/docs/installation-configuration/authentication.
// A route missing from it needs roles.Admin.
var Permissions = apiAuth.Policy{
	// Lets anything check that SOARCA is up.
	route(http.MethodGet, "/status/ping"): {Public: true},
	// The Swagger UI and spec. Calls made from the UI need a token, which
	// the UI's Authorize button adds.
	route(http.MethodGet, "/swagger/*any"): {Public: true},

	route(http.MethodGet, "/status/"): {Permission: roles.PlaybookRead},

	route(http.MethodGet, "/playbook/"):             {Permission: roles.PlaybookRead},
	route(http.MethodGet, "/playbook/meta/"):        {Permission: roles.PlaybookRead},
	route(http.MethodGet, "/playbook/:id"):          {Permission: roles.PlaybookRead},
	route(http.MethodPost, "/playbook/"):            {Permission: roles.PlaybookWrite},
	route(http.MethodPut, "/playbook/:id"):          {Permission: roles.PlaybookWrite},
	route(http.MethodDelete, "/playbook/:id"):       {Permission: roles.PlaybookWrite},
	route(http.MethodGet, "/reporter/"):             {Permission: roles.PlaybookRead},
	route(http.MethodGet, "/reporter/:id"):          {Permission: roles.PlaybookRead},
	route(http.MethodPost, "/trigger/playbook"):     {Permission: roles.PlaybookExecute},
	route(http.MethodPost, "/trigger/playbook/:id"): {Permission: roles.PlaybookExecute},

	route(http.MethodGet, "/manual/"):                  {Permission: roles.PlaybookExecute},
	route(http.MethodGet, "/manual/:exec_id/:step_id"): {Permission: roles.PlaybookExecute},
	route(http.MethodPost, "/manual/continue"):         {Permission: roles.PlaybookExecute},

	route(http.MethodGet, "/keymanagement/"):            {Permission: roles.Admin},
	route(http.MethodPut, "/keymanagement/:keyname"):    {Permission: roles.Admin},
	route(http.MethodPatch, "/keymanagement/:keyname"):  {Permission: roles.Admin},
	route(http.MethodDelete, "/keymanagement/:keyname"): {Permission: roles.Admin},
}

func route(method string, path string) string {
	return apiAuth.RouteKey(method, path)
}
