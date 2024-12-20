package manual

import (
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

// Agent and target for the manual command itself make little sense.
// Unless an agent is the intended system that does post_message, and wait_for_response.
// But the targets? For the automated execution, there is no need to specify any.
//
// It is always either only the internal API, or the internal API and ONE integration for manual.
// Env variable: can only have one active manual interactor.
//
// In light of this, for hierarchical and distributed playbooks executions (via multiple playbook actions),
// 	there will be ONE manual integration (besides internal API) per every ONE SOARCA instance.

// The InteractionCapability manages the manual command infromation and status, like a cache. And interfaces any interactor type (e.g. API, integration)

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
//	@failure		400		{object}	[]manual.InteractionCommandData
//	@Router			/manual/ [GET]
func (manualHandler *ManualHandler) GetPendingCommands(g *gin.Context) {
	commands, status, err := manualHandler.interactionCapability.GetPendingCommands()
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError,
			"Failed get pending manual commands",
			"GET /manual/", err.Error())
		return
	}
	g.JSON(status,
		commands)
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
//	@Success		200		{object}	manual.InteractionCommandData
//	@failure		400		{object}	api.Error
//	@Router			/manual/{exec_id}/{step_id} [GET]
func (manualHandler *ManualHandler) GetPendingCommand(g *gin.Context) {
	execution_id := g.Param("execution_id")
	step_id := g.Param("step_id")
	execId, err := uuid.Parse(execution_id)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to parse execution ID",
			"GET /manual/"+execution_id+"/"+step_id, err.Error())
		return
	}

	executionMetadata := execution.Metadata{ExecutionId: execId, StepId: step_id}
	commandData, status, err := manualHandler.interactionCapability.GetPendingCommand(executionMetadata)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError,
			"Failed to provide pending manual command",
			"GET /manual/"+execution_id+"/"+step_id, err.Error())
		return
	}
	g.JSON(status,
		commandData)
}

// manual
//
//	@Summary		updates the value of a variable according to the manual interaction
//	@Schemes
//	@Description	updates the value of a variable according to the manual interaction
//	@Tags			manual
//	@Accept			json
//	@Produce		json
//	@Param			exec_id				path	string	true	"execution ID"
//	@Param			step_id				path	string	true	"step ID"
//	@Param			type				body	string	true	"type"
//	@Param			outArgs				body	string	true	"execution ID"
//	@Param			execution_id		body	string	true	"playbook ID"
//	@Param			playbook_id			body	string	true	"playbook ID"
//	@Param			step_id				body	string	true	"step ID"
//	@Param			response_status		body	string	true	"response status"
//	@Param			response_out_args	body	manual.ManualOutArgs	true	"out args"
//	@Success		200			{object}	api.Execution
//	@failure		400			{object}	api.Error
//	@Router			/manual/{exec_id}/{step_id} [PATCH]
func (manualHandler *ManualHandler) PatchContinue(g *gin.Context) {

	paramExecutionId := g.Param("exec_id")
	paramStepId := g.Param("step_id")

	jsonData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		log.Error("failed")
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to read json",
			"POST /manual/continue/{exec_id}/{step_id}", "")
		return
	}

	var outArgsUpdate manual.ManualOutArgUpdatePayload
	err = json.Unmarshal(jsonData, &outArgsUpdate)
	if err != nil {
		log.Error("failed to unmarshal JSON")
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Failed to unmarshal JSON",
			"POST /manual/continue/{exec_id}/{step_id}", "")
		return
	}

	if (outArgsUpdate.ExecutionId != paramExecutionId) || (outArgsUpdate.StepId != paramStepId) {
		apiError.SendErrorResponse(g, http.StatusBadRequest,
			"Mismatch between execution ID and step ID between URL parameters and request body",
			"POST /manual/continue/{execution_id}/{step_id}", "")
		return
	}

	status, err := manualHandler.interactionCapability.PostContinue(outArgsUpdate)
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError,
			"Failed to post continue ID",
			"POST /manual/continue", err.Error())
		return
	}
	g.JSON(
		status,
		api.Execution{
			ExecutionId: uuid.MustParse(outArgsUpdate.ExecutionId),
			PlaybookId:  outArgsUpdate.PlaybookId,
		})
}
