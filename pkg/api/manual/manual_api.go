package manual

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/capability/manual/interaction"
	"soarca/pkg/models/api"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apiError "soarca/pkg/api/error"
)

// Notes:
// A manual command in CACAO is simply the operation:
// 		{ post_message; wait_for_response (returning a result) }
// The manual API expose general manual executions wide information
// Thus, we need a ManualHandler that uses an IInteractionStorage, implemented by interactionCapability
// The API routes will invoke the ManualHandler.interactionCapability interface instance
// The InteractionCapability manages the manual command infromation and status, like a cache. And interfaces any interactor type (e.g. API, integration)

// It is always either only the internal API, or the internal API and ONE integration for manual.
// Env variable: can only have one active manual interactor.
//
// In light of this, for hierarchical and distributed playbooks executions (via multiple playbook actions),
// 	there will be ONE manual integration (besides internal API) per every ONE SOARCA instance.

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

type ManualHandler struct {
	interactionCapability interaction.IInteractionStorage
}

func NewManualHandler(interaction interaction.IInteractionStorage) *ManualHandler {
	instance := ManualHandler{}
	instance.interactionCapability = interaction
	return &instance
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
//	@failure		400		{object}	[]api.InteractionCommandData
//	@Router			/manual/ [GET]
func (manualHandler *ManualHandler) GetPendingCommands(g *gin.Context) {
	commands, status, err := manualHandler.interactionCapability.GetPendingCommands()
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError,
			"Failed get pending manual commands",
			"GET /manual/", "")
		return
	}

	response := []api.InteractionCommandData{}
	for _, command := range commands {
		response = append(response, manualHandler.parseCommandInfoToResponse(command))
	}

	g.JSON(status,
		response)
}

// manual
//
//	@Summary		get a specific manual command that still needs a value to be returned
//	@Schemes
//	@Description	get a specific manual command that still needs a value to be returned
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Param			exec_id	path	string	true	"execution ID"
//	@Param			step_id	path	string	true	"step ID"
//	@Success		200		{object}	api.InteractionCommandData
//	@failure		400		{object}	api.Error
//	@Router			/manual/{exec_id}/{step_id} [GET]
func (manualHandler *ManualHandler) GetPendingCommand(g *gin.Context) {
	execution_id := g.Param("exec_id")
	step_id := g.Param("step_id")
	execId, err := uuid.Parse(execution_id)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to parse execution ID",
			"GET /manual/"+execution_id+"/"+step_id, "")
		return
	}

	executionMetadata := execution.Metadata{ExecutionId: execId, StepId: step_id}
	commandData, status, err := manualHandler.interactionCapability.GetPendingCommand(executionMetadata)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError,
			"Failed to provide pending manual command",
			"GET /manual/"+execution_id+"/"+step_id, "")
		return
	}

	commandInfo := manualHandler.parseCommandInfoToResponse(commandData)

	g.JSON(status, commandInfo)
}

// manual
//
//	@Summary		updates the value of a variable according to the manual interaction
//	@Schemes
//	@Description	updates the value of a variable according to the manual interaction
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Param			exec_id				path	string			true	"execution ID"
//	@Param			step_id				path	string			true	"step ID"
//	@Param			type				body	string			true	"type"
//	@Param			outArgs				body	string			true	"execution ID"
//	@Param			execution_id		body	string			true	"playbook ID"
//	@Param			playbook_id			body	string			true	"playbook ID"
//	@Param			step_id				body	string			true	"step ID"
//	@Param			response_status		body	string			true	"response status"
//	@Param			response_out_args	body	cacao.Variables	true	"out args"
//	@Success		200			{object}	api.Execution
//	@failure		400			{object}	api.Error
//	@Router			/manual/continue [POST]
func (manualHandler *ManualHandler) PostContinue(g *gin.Context) {

	jsonData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		log.Error("failed")
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to read json",
			"POST /manual/continue", "")
		return
	}

	var outArgsUpdate api.ManualOutArgsUpdatePayload
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&outArgsUpdate)
	if err != nil {
		log.Error("failed to unmarshal JSON")
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to unmarshal JSON",
			"POST /manual/continue", "")
		return
	}

	for varName, variable := range outArgsUpdate.ResponseOutArgs {
		if varName != variable.Name {
			log.Error("variable name mismatch")
			apiError.SendErrorResponse(g, http.StatusBadRequest,
				"Variable name mismatch",
				"POST /manual/continue", "")
			return
		}
	}

	// Create object to pass to interaction capability
	executionId, err := uuid.Parse(outArgsUpdate.ExecutionId)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to parse execution ID",
			"POST /manual/continue", "")
		return
	}

	interactionResponse := manual.InteractionResponse{
		Metadata: execution.Metadata{
			StepId:      outArgsUpdate.StepId,
			ExecutionId: executionId,
			PlaybookId:  outArgsUpdate.PlaybookId,
		},
		OutArgsVariables: outArgsUpdate.ResponseOutArgs,
		ResponseStatus:   outArgsUpdate.ResponseStatus,
		ResponseError:    nil,
	}

	status, err := manualHandler.interactionCapability.PostContinue(interactionResponse)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError,
			"Failed to post continue ID",
			"POST /manual/continue", "")
		return
	}
	g.JSON(
		status,
		api.Execution{
			ExecutionId: uuid.MustParse(outArgsUpdate.ExecutionId),
			PlaybookId:  outArgsUpdate.PlaybookId,
		})
}

// Utility

func (manualHandler *ManualHandler) parseCommandInfoToResponse(commandInfo manual.CommandInfo) api.InteractionCommandData {
	commandText := commandInfo.Context.Command.Command
	isBase64 := false
	if len(commandInfo.Context.Command.CommandB64) > 0 {
		commandText = commandInfo.Context.Command.CommandB64
		isBase64 = true
	}

	response := api.InteractionCommandData{
		Type:            "manual-command-info",
		ExecutionId:     commandInfo.Metadata.ExecutionId.String(),
		PlaybookId:      commandInfo.Metadata.PlaybookId,
		StepId:          commandInfo.Metadata.StepId,
		Description:     commandInfo.Context.Command.Description,
		Command:         commandText,
		CommandIsBase64: isBase64,
		Target:          commandInfo.Context.Target,
		OutVariables:    commandInfo.OutArgsVariables,
	}

	return response
}
