package status

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Helloworld(g *gin.Context) {
	g.JSON(http.StatusOK, "helloworld from /step")
}

// GET     /step

func Routes(route *gin.Engine){
	coa := route.Group("/step")
	{
		coa.GET("/", Helloworld)
		
		//workflow.POST()
	}
	
}