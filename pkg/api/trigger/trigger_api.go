package trigger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/logger"
	"soarca/pkg/core/decomposer"
	"soarca/pkg/models/api"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/decoder"
	"time"

	apiError "soarca/pkg/api/error"

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
	controller        decomposer_controller.IController
	database          database.IController
	ExecutionsChannel chan decomposer.ExecutionDetails
}

func NewTriggerHandler(controller decomposer_controller.IController, database database.IController) *TriggerHandler {
	instance := TriggerHandler{}
	instance.controller = controller
	instance.database = database
	// Channel to get back execution details
	instance.ExecutionsChannel = make(chan decomposer.ExecutionDetails)
	return &instance
}

// trigger
//
//	@Summary	trigger a playbook by id that is stored in SOARCA
//	@Schemes
//	@Description	trigger playbook by id
//	@Tags			trigger
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"playbook ID"
//	@Param			data	body		cacao.Variables	true	"playbook"
//	@Success		200		{object}	api.Execution
//	@failure		400		{object}	api.Error
//	@Router			/trigger/playbook/{id} [POST]
func (handler *TriggerHandler) ExecuteById(context *gin.Context) {
	log.Trace("received execute by ID")
	id := context.Param("id")

	db := handler.database.GetDatabaseInstance()
	playbook, err := db.Read(id)
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
	decomposer := handler.controller.NewDecomposer()
	go decomposer.ExecuteAsync(*playbook, handler.ExecutionsChannel)
	timer := time.NewTimer(time.Duration(3) * time.Second)
	for {
		select {
		case <-timer.C:
			log.Error("async execution timed out for playbook ", playbook.ID)

			apiError.SendErrorResponse(context,
				http.StatusRequestTimeout,
				"async execution timed out for playbook "+playbook.ID,
				"POST "+context.Request.URL.Path, "")
			return

		case executionsDetail := <-handler.ExecutionsChannel:
			playbookId := executionsDetail.PlaybookId
			executionId := executionsDetail.ExecutionId
			if playbookId == playbook.ID {
				context.JSON(http.StatusOK,
					api.Execution{
						ExecutionId: executionId,
						PlaybookId:  playbookId,
					})
				return
			}
		}
	}
}

// public fun as tested externally (integration test)
func MergeVariablesInPlaybook(playbook *cacao.Playbook, body []byte) error {
	payloadVariables := cacao.NewVariables()
	err := json.Unmarshal(body, &payloadVariables)
	if err != nil {
		log.Trace(err)
		return errors.New("cannot unmarshal provided variables")
	}

	// Check payload-injected variables are valid set for playbook variables
	for name, variable := range payloadVariables {
		// Must exist
		if _, ok := playbook.PlaybookVariables[name]; !ok {
			return fmt.Errorf("provided variables is not a valid subset of the variables for the referenced playbook [ playbook id: %s ]", playbook.ID)
		}
		// Exists, playbook var type must match
		if variable.Type != playbook.PlaybookVariables[name].Type {
			return fmt.Errorf("mismatch in variables type for [ %s ]: payload var type = %s, playbook var type = %s", name, variable.Type, playbook.PlaybookVariables[name].Type)
		}
		// Exists, playbook var must be external
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
