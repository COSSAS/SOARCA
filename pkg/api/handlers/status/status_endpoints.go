package status

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// /Status/ping GET handler for handling status api calls
// Returns the status model object for SOARCA
//
//	@Summary	ping to see if SOARCA is up returns pong
//	@Schemes
//	@Description	return SOARCA status
//	@Tags			ping pong
//	@Produce		plain
//	@success		200	string	pong
//	@Router			/status/ping [GET]
func Pong(g *gin.Context) {
	g.Data(http.StatusOK, "text/plain", []byte("pong"))
}

// GET     /status
// GET     /status/ping
func Routes(route *gin.Engine) {
	router := route.Group("/status")
	{
		router.GET("/", Api)
		router.GET("/ping", Pong)

	}
}
