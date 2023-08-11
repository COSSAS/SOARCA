package workflow

import (
	"net/http"
	// cacao "soarca/internal/cacao"

	"github.com/gin-gonic/gin"
)


func Helloworld(g *gin.Context) {
	g.JSON(http.StatusOK, "helloworld from /workdlow")
}

// GET     /workflow
// POST    /workflow
// GET     /workflow/workflow-id
// PUT     /workflow/workflow-id
// DELETE  /workflow/workflow-id

func Routes(route *gin.Engine){
	workflow := route.Group("/workflow")
	{
		workflow.GET("/", Helloworld)
		workflow.POST("/", SubmitWorkflow)
	}

}