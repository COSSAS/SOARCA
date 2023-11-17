package trigger

import (
	"github.com/gin-gonic/gin"
)

// POST    /operator/coa/coa-id
func Routes(route *gin.Engine, trigger *TriggerApi) {
	group := route.Group("/trigger")
	{
		group.POST("/workflow", trigger.Execute)
	}

}
