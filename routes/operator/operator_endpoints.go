package status

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Helloworld(g *gin.Context) {
	g.JSON(http.StatusOK, "helloworld from /operator")
}

//POST    /operator/coa/coa-id
func Routes(route *gin.Engine){
	coa := route.Group("/operator")
	{
		coa.POST("/coa/:coa-id", Helloworld)
		//workflow.POST()
	}
	
}