package coa

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Helloworld(g *gin.Context) {
	g.JSON(http.StatusOK, "helloworld from /coa")
}


func id_tester(g *gin.Context){
id := g.Param("coa-id")
fmt.Println(id)

}

// GET     /coa
// POST    /coa
// GET     /coa/coa-id
// PUT     /coa/coa-id
// DELETE  /coa/coa-id

func Routes(route *gin.Engine){
	coa := route.Group("/coa")
	{
		
		coa.GET("/",Helloworld)
		coa.POST("/:coa-id", id_tester)
		coa.PUT("/:coa-id", id_tester)
		coa.DELETE("/:coa-id", id_tester)
	}
	
}