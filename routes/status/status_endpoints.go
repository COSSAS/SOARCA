package coa

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Helloworld(g *gin.Context) {
	g.JSON(http.StatusOK, "helloworld from /status")
}

func id_tester(g *gin.Context){

	
		// Get the value of the 'id' parameter from the URL
	id := g.Param("id")
	fmt.Println(id)

}
// GET     /status
// GET     /status/workflow
// GET     /status/workflow/id
// GET     /status/coa/id
// GET     /status/history
func Routes(route *gin.Engine){
	coa := route.Group("/status")
	{
		coa.GET("/", Helloworld)
		coa.GET("/workflow/:id", id_tester)
		coa.GET("/coa/:id",id_tester)
		coa.GET("/history",Helloworld)
		//workflow.POST()
	}
	
}