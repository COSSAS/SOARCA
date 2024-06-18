package status

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
