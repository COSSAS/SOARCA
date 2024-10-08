package trigger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/decomposer"
	"soarca/logger"
	"soarca/models/api"
	"soarca/models/cacao"
	"soarca/models/decoder"
	apiError "soarca/routes/error"

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

type TriggerApi struct {
	controller        decomposer_controller.IController
	database          database.IController
	ExecutionsChannel chan decomposer.ExecutionDetails
}

func New(controller decomposer_controller.IController, database database.IController) *TriggerApi {
	instance := TriggerApi{}
	instance.controller = controller
	instance.database = database
	// Channel to get back execution details
	instance.ExecutionsChannel = make(chan decomposer.ExecutionDetails)
	return &instance
}

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
func (trigger *TriggerApi) ExecuteById(context *gin.Context) {
	id := context.Param("id")

	db := trigger.database.GetDatabaseInstance()
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
			apiError.SendErrorResponse(context, http.StatusBadRequest, fmt.Sprintf("Cannot execute. reason: %s", err), "POST /trigger/playbook/"+id, "")
			return
		}
	}
	trigger.execute(&playbook, context)
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
func (trigger *TriggerApi) Execute(context *gin.Context) {

	jsonData, err := io.ReadAll(context.Request.Body)
	if err != nil {
		log.Error("failed")
		apiError.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to marshall json on server side",
			"POST /trigger/playbook", "")
		return
	}
	// playbook := cacao.Decode(jsonData)
	playbook := decoder.DecodeValidate(jsonData)
	if playbook == nil {
		apiError.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to decode playbook",
			"POST /trigger/playbook", "")
		return
	}

	trigger.execute(playbook, context)
}

func (trigger *TriggerApi) execute(playbook *cacao.Playbook, context *gin.Context) {
	decomposer := trigger.controller.NewDecomposer()
	go decomposer.ExecuteAsync(*playbook, trigger.ExecutionsChannel)
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

		case executionsDetail := <-trigger.ExecutionsChannel:
			playbookId := executionsDetail.PlaybookId
			executionId := executionsDetail.ExecutionId
			if playbookId == playbook.ID {
				context.JSON(http.StatusOK,
					api.Execution{ExecutionId: executionId,
						PlaybookId: playbookId})
				return
			}
		}
	}
}
