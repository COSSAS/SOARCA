package trigger

import (
	"github.com/gin-gonic/gin"
)

// trigger
// @Summary trigger a workflow with via cacao payload
// @Schemes
// @Description trigger workflow
// @Tags workflow
// @Accept json
// @Produce json
// @Param  playbook body cacao.Playbook true "execute playbook by payload"
// @Success 200 "{"execution_id":"uuid","payload":"playbook--uuid"}"
// @error	400
// @Router /trigger/workflow [POST]
func Routes(route *gin.Engine, trigger *TriggerApi) {
	group := route.Group("/trigger")
	{
		group.POST("/workflow", trigger.Execute)
	}

}
