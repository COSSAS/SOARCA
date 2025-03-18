package manual

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	return &ManualHandler{interactionCapability: interaction}
}

// manual
//
//	@Summary	get all pending manual commands that still needs values to be returned
//	@Schemes
//	@Description	get all pending manual commands that still needs values to be returned
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	api.Execution
//	@failure		400	{object}	[]api.InteractionCommandData
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
//	@Summary	get a specific manual command that still needs a value to be returned
//	@Schemes
//	@Description	get a specific manual command that still needs a value to be returned
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Param			exec_id	path		string	true	"execution ID"
//	@Param			step_id	path		string	true	"step ID"
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
//	@Summary	updates the value of a variable according to the manual interaction
//	@Schemes
//	@Description	updates the value of a variable according to the manual interaction
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Param			exec_id	path		string							true	"execution ID"
//	@Param			step_id	path		string							true	"step ID"
//	@Param			data	body		api.ManualOutArgsUpdatePayload	true	"playbook"
//	@Success		200		{object}	api.Execution
//	@failure		400		{object}	api.Error
//	@Router			/manual/continue [POST]
func (manualHandler *ManualHandler) PostContinue(g *gin.Context) {

	byteData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		log.Error("failed")
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to read json",
			"POST /manual/continue", "")
		return
	}

	outArgsUpdate, err := manualHandler.parseManualOutArgsUpdate(byteData)
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			fmt.Sprint(fmt.Errorf("Failed to parse manual out args payload: %w", err)),
			"POST /manual/continue", err.Error())
		return
	}

	interactionResponse, err := manualHandler.parseManualOutArgsToInteractionResponse(outArgsUpdate)
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to parse response",
			"POST /manual/continue", err.Error())
		return
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
		return
	}

	g.JSON(
		http.StatusOK,
		api.Execution{
			ExecutionId: executionId,
			PlaybookId:  outArgsUpdate.PlaybookId,
		})
}

// ############################################################################
// Utility
// ############################################################################

func (manualHandler *ManualHandler) parseManualOutArgsUpdate(postData []byte) (api.ManualOutArgsUpdatePayload, error) {
	decoder := json.NewDecoder(bytes.NewReader(postData))
	decoder.DisallowUnknownFields()
	var outArgsUpdate api.ManualOutArgsUpdatePayload
	err := decoder.Decode(&outArgsUpdate)
	if err != nil {
		errorString := fmt.Errorf("failed to unmarshal JSON: %w", err)
		log.Error(errorString)
		return api.ManualOutArgsUpdatePayload{}, errorString
	}

	// Check if variable names match
	for varName, variable := range outArgsUpdate.ResponseOutArgs {
		if varName != variable.Name {
			errorString := fmt.Errorf(
				"variable name mismatch for variable %s: has different name property: %s",
				varName, variable.Name)
			log.Error(errorString)
			return api.ManualOutArgsUpdatePayload{}, errorString
		}
	}

	return outArgsUpdate, nil
}

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

func (manualHandler *ManualHandler) parseManualOutArgsToInteractionResponse(response api.ManualOutArgsUpdatePayload) (manual.InteractionResponse, error) {
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
