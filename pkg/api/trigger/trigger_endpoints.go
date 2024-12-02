package trigger

import (
	"github.com/gin-gonic/gin"
)

func Routes(route *gin.Engine, trigger *TriggerApi) {
	group := route.Group("/trigger")
	{
		group.POST("/playbook", trigger.Execute)
		group.POST("/playbook/:id", trigger.ExecuteById)
	}
}
