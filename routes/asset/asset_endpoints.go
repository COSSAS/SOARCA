package status

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Helloworld(g *gin.Context) {
	g.JSON(http.StatusOK, "helloworld from /asset")
}

func id_tester(g *gin.Context){

		// Get the value of the 'id' parameter from the URL
	id := g.Param("id")
	fmt.Println(id)
}
// GET     /asset
// POST    /asset
// GET     /asset/id
// PUT     /asset/id
// DELETE  /asset/id
func Routes(route *gin.Engine){
	coa := route.Group("/asset")
	{
		coa.GET("/", Helloworld)
		coa.POST("/", Helloworld)
		coa.GET("/:id",id_tester)
		coa.PUT("/:id",id_tester)
		coa.DELETE("/:id", id_tester)


		//workflow.POST()
	}
	
}