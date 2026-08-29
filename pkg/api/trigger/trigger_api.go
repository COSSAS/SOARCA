package trigger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/internal/services"
	"soarca/internal/storage"
	apiError "soarca/pkg/api/error"
	"soarca/pkg/models/api"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/decoder"

	"github.com/gin-gonic/gin"
)

type Empty struct{}

var log *logger.Log

type ITrigger interface {
	Execute(context *gin.Context)
}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

type TriggerHandler struct {
	executionRuntime services.ExecutionRuntime
	playbookStore    storage.PlaybookStore
}

func NewTriggerHandler(executionRuntime services.ExecutionRuntime, playbookStore storage.PlaybookStore) *TriggerHandler {
	return &TriggerHandler{
		executionRuntime: executionRuntime,
		playbookStore:    playbookStore,
	}
}

func (handler *TriggerHandler) ExecuteById(context *gin.Context) {
	log.Trace("received execute by ID")
	id := context.Param("id")

	playbook, err := handler.playbookStore.Get(context.Request.Context(), id)
	if err != nil {
		log.Error("failed to load playbook")
		apiError.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to load playbook",
			"POST /trigger/playbook/"+id, err.Error())
		return
	}
	if context.Request.Body != nil {
		jsonData, err := io.ReadAll(context.Request.Body)
		if err != nil {
			log.Trace("Playbook trigger has failed to decode request body")
			apiError.SendErrorResponse(context, http.StatusBadRequest, "Failed to decode request body", "POST /trigger/playbook/"+id, "")
			return
		}
		err = MergeVariablesInPlaybook(&playbook, jsonData)
		if err != nil {
			log.Error(err)
			apiError.SendErrorResponse(context, http.StatusBadRequest, fmt.Sprintf("Cannot execute. reason: %s", err), "POST /trigger/playbook/"+id, "")
			return
		}
	}
	handler.executePlaybook(&playbook, context)
}

func (handler *TriggerHandler) Execute(context *gin.Context) {
	log.Trace("received execute with body")
	jsonData, err := io.ReadAll(context.Request.Body)
	if err != nil {
		log.Error("failed")
		apiError.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to marshall json on server side",
			"POST /trigger/playbook", "")
		return
	}
	playbook := decoder.DecodeValidate(jsonData)
	if playbook == nil {
		log.Error("Failed to decode playbook")
		apiError.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to decode playbook",
			"POST /trigger/playbook", "")
		return
	}

	handler.executePlaybook(playbook, context)
}

func (handler *TriggerHandler) executePlaybook(playbook *cacao.Playbook, context *gin.Context) {
	executionId, err := handler.executionRuntime.StartExecution(context.Request.Context(), playbook, cacao.Variables{})
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(context,
			http.StatusRequestTimeout,
			err.Error(),
			"POST "+context.Request.URL.Path, "")
		return
	}
	context.JSON(http.StatusOK,
		api.Execution{
			ExecutionId: executionId,
			PlaybookId:  playbook.ID,
		})
}

func MergeVariablesInPlaybook(playbook *cacao.Playbook, body []byte) error {
	payloadVariables := cacao.NewVariables()
	err := json.Unmarshal(body, &payloadVariables)
	if err != nil {
		log.Trace(err)
		return errors.New("cannot unmarshal provided variables")
	}
	for name, variable := range payloadVariables {
		if _, ok := playbook.PlaybookVariables[name]; !ok {
			return fmt.Errorf("provided variables is not a valid subset of the variables for the referenced playbook [ playbook id: %s ]", playbook.ID)
		}
		if variable.Type != playbook.PlaybookVariables[name].Type {
			return fmt.Errorf("mismatch in variables type for [ %s ]: payload var type = %s, playbook var type = %s", name, variable.Type, playbook.PlaybookVariables[name].Type)
		}
		if !playbook.PlaybookVariables[name].External {
			return fmt.Errorf("playbook variable [ %s ] cannot be assigned in playbook because it is not marked as external in the plabook", name)
		}

		updatedVariable := cacao.Variable{
			Name:        name,
			Type:        playbook.PlaybookVariables[name].Type,
			Description: playbook.PlaybookVariables[name].Description,
			Value:       variable.Value,
			Constant:    playbook.PlaybookVariables[name].Constant,
			External:    playbook.PlaybookVariables[name].External,
		}
		playbook.PlaybookVariables[name] = updatedVariable
	}
	return nil
}
