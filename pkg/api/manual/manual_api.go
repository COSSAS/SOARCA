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
	"soarca/internal/services"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/api"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apiError "soarca/pkg/api/error"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

type ManualHandler struct {
	inbox services.ManualInbox
}

func NewManualHandler(inbox services.ManualInbox) *ManualHandler {
	return &ManualHandler{inbox: inbox}
}

// manual
//
//	@Summary	get all pending manual commands that still needs values to be returned
//	@Schemes
//	@Description	get all pending manual commands that still needs values to be returned
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	[]api.InteractionCommandData
//	@failure		400	{object}	[]api.InteractionCommandData
//	@Router			/manual/ [GET]
func (manualHandler *ManualHandler) GetPendingCommands(g *gin.Context) {
	commands, err := manualHandler.inbox.ListPendingCommands()
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
//	@Param			exec_id				path		string	true	"execution ID"
//	@Param			step_execution_id	path		string	true	"step execution ID (identifies a specific pending step invocation; see GET /manual/ to discover it, as multiple pending commands may share the same step ID)"
//	@Success		200					{object}	api.InteractionCommandData
//	@failure		400					{object}	api.Error
//	@Router			/manual/{exec_id}/{step_execution_id} [GET]
func (manualHandler *ManualHandler) GetPendingCommand(g *gin.Context) {
	execution_id := g.Param("exec_id")
	step_execution_id := g.Param("step_execution_id")
	execId, err := uuid.Parse(execution_id)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to parse execution ID",
			"GET /manual/"+execution_id+"/"+step_execution_id, "")
		return
	}
	stepExecutionId, err := uuid.Parse(step_execution_id)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to parse step execution ID",
			"GET /manual/"+execution_id+"/"+step_execution_id, "")
		return
	}

	commandData, err := manualHandler.inbox.GetPendingCommand(execution.Metadata{ExecutionId: execId, StepExecutionId: stepExecutionId})
	if err != nil {
		log.Error(err)
		code := http.StatusBadRequest
		if errors.Is(err, manual.ErrorPendingCommandNotFound{}) {
			code = http.StatusNotFound
		}
		apiError.SendErrorResponse(g, code,
			"Failed to provide pending manual command",
			"GET /manual/"+execution_id+"/"+step_execution_id, "")
		return
	}

	commandInfo := manualHandler.parseCommandInfoToResponse(commandData)

	g.JSON(http.StatusOK, commandInfo)
}

// manual
//
//	@Summary	resolve a specific pending manual command by supplying its out args
//	@Schemes
//	@Description	resolve a specific pending manual command by supplying its out args. This is a PUT
//	@Description	on the same resource GET /manual/{exec_id}/{step_execution_id} identifies, not a
//	@Description	generic RPC-style action, so the ids live in the path, not the body.
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Param			exec_id				path		string							true	"execution ID"
//	@Param			step_execution_id	path		string							true	"step execution ID (identifies a specific pending step invocation; see GET /manual/ to discover it, as multiple pending commands may share the same step ID)"
//	@Param			data				body		api.ManualOutArgsUpdatePayload	true	"resolution"
//	@Success		200					{object}	api.Execution
//	@failure		400					{object}	api.Error
//	@Router			/manual/{exec_id}/{step_execution_id} [PUT]
func (manualHandler *ManualHandler) PutContinue(g *gin.Context) {
	execution_id := g.Param("exec_id")
	step_execution_id := g.Param("step_execution_id")
	route := "PUT /manual/" + execution_id + "/" + step_execution_id

	execId, err := uuid.Parse(execution_id)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to parse execution ID",
			route, "")
		return
	}
	stepExecutionId, err := uuid.Parse(step_execution_id)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to parse step execution ID",
			route, "")
		return
	}
	byteData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		log.Error("failed")
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to read json",
			route, "")
		return
	}

	outArgsUpdate, err := manualHandler.parseManualOutArgsUpdate(byteData)
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			fmt.Sprint(fmt.Errorf("failed to parse manual out args payload: %w", err)),
			route, err.Error())
		return
	}

	// Looked up here (rather than only implicitly inside PostContinue below)
	// so the response can report the PlaybookId, and so an unknown resource
	// is reported before any out-args validation runs against it.
	pendingCommand, err := manualHandler.inbox.GetPendingCommand(execution.Metadata{ExecutionId: execId, StepExecutionId: stepExecutionId})
	if err != nil {
		log.Error(err)
		code := http.StatusBadRequest
		if errors.Is(err, manual.ErrorPendingCommandNotFound{}) {
			code = http.StatusNotFound
		}
		apiError.SendErrorResponse(g, code,
			"Pending manual command not found",
			route, "")
		return
	}

	interactionResponse := manualHandler.parseManualOutArgsToInteractionResponse(pendingCommand.Metadata, outArgsUpdate)

	err = manualHandler.inbox.ContinuePendingCommand(interactionResponse)
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
			route, "")
		return
	}

	g.JSON(
		http.StatusOK,
		api.Execution{
			ExecutionId: execId,
			PlaybookId:  pendingCommand.Metadata.PlaybookId,
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
	// Manual is a human-resolved, single-outcome step (one response resolves
	// the whole pending entry), but a step may list multiple commands and
	// targets -- surface all of them rather than only the first. Multiple
	// pending commands may share the same StepId (e.g. overlapping
	// while-loop iterations); each is a distinct entry here, disambiguated
	// by StepExecutionId.
	commands := make([]api.ManualCommand, 0, len(commandInfo.Context.Commands))
	for _, command := range commandInfo.Context.Commands {
		commandText := command.Command
		isBase64 := false
		if len(command.CommandB64) > 0 {
			commandText = command.CommandB64
			isBase64 = true
		}
		commands = append(commands, api.ManualCommand{
			Description:     command.Description,
			Command:         commandText,
			CommandIsBase64: isBase64,
		})
	}

	targets := make([]capability.ResolvedTarget, 0, len(commandInfo.Context.Targets))
	for _, resolvedTarget := range commandInfo.Context.Targets {
		targets = append(targets, capability.ResolvedTarget{
			Target:         resolvedTarget.Target,
			Authentication: resolvedTarget.Authentication,
		})
	}

	response := api.InteractionCommandData{
		Type:            "manual-command-info",
		ExecutionId:     commandInfo.Metadata.ExecutionId.String(),
		PlaybookId:      commandInfo.Metadata.PlaybookId,
		StepId:          commandInfo.Metadata.StepId,
		StepExecutionId: commandInfo.Metadata.StepExecutionId.String(),
		Commands:        commands,
		Targets:         targets,
		OutVariables:    commandInfo.OutArgsVariables,
	}

	return response
}

func (manualHandler *ManualHandler) parseManualOutArgsToInteractionResponse(
	metadata execution.Metadata,
	response api.ManualOutArgsUpdatePayload,
) manual.InteractionResponse {
	return manual.InteractionResponse{
		Metadata:         metadata,
		ResponseStatus:   response.ResponseStatus,
		OutArgsVariables: response.ResponseOutArgs,
		ResponseError:    nil,
	}
}
