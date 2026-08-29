package trigger

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/internal/services"
	triggersvc "soarca/internal/services/trigger"
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
	triggerService services.TriggerService
}

func NewTriggerHandler(triggerService services.TriggerService) *TriggerHandler {
	return &TriggerHandler{triggerService: triggerService}
}

func (handler *TriggerHandler) ExecuteById(context *gin.Context) {
	log.Trace("received execute by ID")
	id := context.Param("id")
	var variables cacao.Variables
	if context.Request.Body != nil {
		jsonData, err := io.ReadAll(context.Request.Body)
		if err != nil {
			log.Trace("Playbook trigger has failed to decode request body")
			apiError.SendErrorResponse(context, http.StatusBadRequest, "Failed to decode request body", "POST /trigger/playbook/"+id, "")
			return
		}
		variables, err = triggersvc.DecodeVariables(jsonData)
		if err != nil {
			log.Error(err)
			apiError.SendErrorResponse(context, http.StatusBadRequest, fmt.Sprintf("Cannot execute. reason: %s", err), "POST /trigger/playbook/"+id, "")
			return
		}
	}
	executionID, err := handler.triggerService.ExecutePlaybook(context.Request.Context(), id, variables)
	if err != nil {
		log.Error(err)
		var validationErr triggersvc.ValidationError
		if errors.As(err, &validationErr) {
			apiError.SendErrorResponse(context, http.StatusBadRequest, fmt.Sprintf("Cannot execute. reason: %s", validationErr.Error()), "POST /trigger/playbook/"+id, "")
			return
		}
		apiError.SendErrorResponse(context, http.StatusRequestTimeout, err.Error(), "POST "+context.Request.URL.Path, "")
		return
	}
	context.JSON(http.StatusOK, api.Execution{
		ExecutionId: executionID,
		PlaybookId:  id,
	})
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

	executionID, err := handler.triggerService.ExecuteUploadedPlaybook(context.Request.Context(), playbook, cacao.Variables{})
	if err != nil {
		log.Error(err)
		var validationErr triggersvc.ValidationError
		if errors.As(err, &validationErr) {
			apiError.SendErrorResponse(context, http.StatusBadRequest, fmt.Sprintf("Cannot execute. reason: %s", validationErr.Error()), "POST /trigger/playbook", "")
			return
		}
		apiError.SendErrorResponse(context,
			http.StatusRequestTimeout,
			err.Error(),
			"POST "+context.Request.URL.Path, "")
		return
	}
	context.JSON(http.StatusOK,
		api.Execution{
			ExecutionId: executionID,
			PlaybookId:  playbook.ID,
		})
}
