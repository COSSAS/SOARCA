package manual

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/capability/manual/interaction"
	"soarca/pkg/models/api"
	"soarca/pkg/models/cacao"
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
	return &ManualHandler{interactionCapability: interaction}
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
	commands, err := manualHandler.interactionCapability.GetPendingCommands()
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

	g.JSON(http.StatusOK,
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
	commandData, err := manualHandler.interactionCapability.GetPendingCommand(executionMetadata)
	if err != nil {
		log.Error(err)
		code := http.StatusBadRequest
		if errors.Is(err, manual.ErrorPendingCommandNotFound{}) {
			code = http.StatusNotFound
		}
		apiError.SendErrorResponse(g, code,
			"Failed to provide pending manual command",
			"GET /manual/"+execution_id+"/"+step_id, "")
		return
	}

	commandInfo := manualHandler.parseCommandInfoToResponse(commandData)

	g.JSON(http.StatusOK, commandInfo)
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

	// Check if variable names match
	ok := manualHandler.postContinueVariableNamesMatchCheck(outArgsUpdate.ResponseOutArgs)
	if !ok {
		log.Error("variable name mismatch")
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Variable name mismatch",
			"POST /manual/continue", "")
		return
	}

	interactionResponse, err := manualHandler.parseManualResponseToInteractionResponse(outArgsUpdate)
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to parse response",
			"POST /manual/continue", err.Error())
	}

	err = manualHandler.interactionCapability.PostContinue(interactionResponse)
	if err != nil {
		log.Error(err)
		code := http.StatusBadRequest
		msg := "Failed to post the continue request"
		if errors.Is(err, manual.ErrorPendingCommandNotFound{}) {
			code = http.StatusNotFound
			msg = "Pending command not found"
		} else if errors.Is(err, manual.ErrorNonMatchingOutArgs{}) {
			code = http.StatusBadRequest
			msg = "Provided out args don't match with expected"
		}
		apiError.SendErrorResponse(g, code,
			msg,
			"POST /manual/continue", "")
		return
	}
	executionId, err := uuid.Parse(outArgsUpdate.ExecutionId)
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusInternalServerError,
			"Failed to parse execution ID",
			"POST /manual/continue", "")
	}

	g.JSON(
		http.StatusOK,
		api.Execution{
			ExecutionId: executionId,
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

func (manualHandler *ManualHandler) parseManualResponseToInteractionResponse(response api.ManualOutArgsUpdatePayload) (manual.InteractionResponse, error) {
	executionId, err := uuid.Parse(response.ExecutionId)
	if err != nil {
		return manual.InteractionResponse{}, err
	}

	interactionResponse := manual.InteractionResponse{
		Metadata: execution.Metadata{
			ExecutionId: executionId,
			PlaybookId:  response.PlaybookId,
			StepId:      response.StepId,
		},
		ResponseStatus:   response.ResponseStatus,
		OutArgsVariables: response.ResponseOutArgs,
		ResponseError:    nil,
	}

	return interactionResponse, nil
}

func (ManualHandler *ManualHandler) postContinueVariableNamesMatchCheck(outArgs cacao.Variables) bool {
	ok := true
	for varName, variable := range outArgs {
		if varName != variable.Name {
			ok = false
		}
	}
	return ok
}
