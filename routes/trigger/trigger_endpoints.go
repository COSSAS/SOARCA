package trigger

import (
	"github.com/gin-gonic/gin"
)

// trigger
//
//	@Summary	trigger a playbook by supplying a cacao playbook payload
//	@Schemes
//	@Description	trigger playbook
//	@Tags			trigger
//	@Accept			json
//	@Produce		json
//	@Param			playbook	body		cacao.Playbook	true	"execute playbook by payload"
//	@Success		200			{object}	api.Execution
//	@failure		400			{object}	api.Error
//	@Router			/trigger/playbook [POST]
func Routes(route *gin.Engine, trigger *TriggerApi) {
	group := route.Group("/trigger")
	{
		group.POST("/playbook", trigger.Execute)
	}
}
