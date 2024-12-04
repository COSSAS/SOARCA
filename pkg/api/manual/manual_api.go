package manual

import (
	"github.com/gin-gonic/gin"
)

type ManualHandler struct {
}

// manual
//
//	@Summary		get all pending manual commands that still needs values to be returned
//	@Schemes
//	@Description	get all pending manual commands that still needs values to be returned
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	api.Execution
//	@failure		400		{object}	api.Error
//	@Router			/manual/ [GET]
func (manualHandler *ManualHandler) GetPendingCommands(g *gin.Context) {

}

// manual
//
//	@Summary		get a specific manual command that still needs a value to be returned
//	@Schemes
//	@Description	get a specific manual command that still needs a value to be returned
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Param			execution_id	path	string	true	"execution ID"
//	@Param			step_id			path	string	true	"step ID"
//	@Success		200		{object}	api.Execution
//	@failure		400		{object}	api.Error
//	@Router			/manual/{execution_id}/{step_id} [GET]
func (manualHandler *ManualHandler) GetPendingCommand(g *gin.Context) {

}

// manual
//
//	@Summary		updates the value of a variable according to the manual interaction
//	@Schemes
//	@Description	updates the value of a variable according to the manual interaction
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Param			type	body		string	true	"type"
//	@Param			execution_id		body	string	true	"execution ID"
//	@Param			playbook_id			body	string	true	"playbook ID"
//	@Param			step_id				body	string	true	"step ID"
//	@Param			response_status		body	string	true	"response status"
//	@Param			response_out_args	body	model.ResponseOutArgs	true	"out args"
//	@Success		200			{object}	api.Execution
//	@failure		400			{object}	api.Error
//	@Router			/manual/continue/ [POST]
func (manualHandler *ManualHandler) PostContinue(g *gin.Context) {

}
