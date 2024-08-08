package trigger

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/decomposer"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/decoder"
	"soarca/routes/error"

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
	controller   decomposer_controller.IController
	database     database.IController
	Executionsch chan decomposer.ExecutionDetails
}

func New(controller decomposer_controller.IController, database database.IController) *TriggerApi {
	instance := TriggerApi{}
	instance.controller = controller
	instance.database = database
	// Channel to get back execution details
	instance.Executionsch = make(chan decomposer.ExecutionDetails)
	return &instance
}

func MergeVariablesInPlaybook(playbook *cacao.Playbook, body []byte) (bool, string) {

	payloadVariables := cacao.NewVariables()
	err := json.Unmarshal(body, &payloadVariables)
	if err != nil {
		log.Trace(err)
		return false, "cannot unmarshal provided variables"
	}

	// Check payload-injected variables are valid set for playbook variables
	for name, variable := range payloadVariables {
		// Must exist
		if _, ok := playbook.PlaybookVariables[name]; !ok {
			return false, fmt.Sprintf("provided variables is not a valid subset of the variables for the referenced playbook [ playbook id: %s ]", playbook.ID)
		}
		// Exists, playbook var type must match
		if variable.Type != playbook.PlaybookVariables[name].Type {
			return false, fmt.Sprintf("mismatch in variables type for [ %s ]: payload var type = %s, playbook var type = %s", name, variable.Type, playbook.PlaybookVariables[name].Type)
		}
		// Exists, playbook var must be external
		if !playbook.PlaybookVariables[name].External {
			return false, fmt.Sprintf("playbook variable [ %s ] cannot be assigned in playbook because it is not marked as external in the plabook", name)
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
	return true, ""
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
		error.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to load playbook",
			"POST /trigger/playbook/"+id, err.Error())
		return
	}
	if context.Request.Body != nil {
		jsonData, err := io.ReadAll(context.Request.Body)
		if err != nil {
			log.Trace("Playbook trigger has failed to decode request body")
			error.SendErrorResponse(context, http.StatusBadRequest, "Failed to decode request body", "POST /trigger/playbook/"+id, "")
		}
		ok, str := MergeVariablesInPlaybook(&playbook, jsonData)
		if !ok {
			error.SendErrorResponse(context, http.StatusBadRequest, fmt.Sprintf("Cannot execute. reason: %s", str), "POST /trigger/playbook/"+id, "")
			return
		}
	}
	fmt.Println(playbook)
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

	jsonData, errIo := io.ReadAll(context.Request.Body)
	if errIo != nil {
		log.Error("failed")
		error.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to marshall json on server side",
			"POST /trigger/playbook", "")
		return
	}
	// playbook := cacao.Decode(jsonData)
	playbook := decoder.DecodeValidate(jsonData)
	if playbook == nil {
		error.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to decode playbook",
			"POST /trigger/playbook", "")
		return
	}

	trigger.execute(playbook, context)
}

func (trigger *TriggerApi) execute(playbook *cacao.Playbook, context *gin.Context) {
	decomposer := trigger.controller.NewDecomposer()
	go decomposer.ExecuteAsync(*playbook, trigger.Executionsch)
	timer := time.NewTimer(time.Duration(3) * time.Second)
	for {
		select {
		case <-timer.C:
			msg := gin.H{
				"execution_id": nil,
				"payload":      playbook.ID,
			}
			context.JSON(http.StatusRequestTimeout, msg)
			log.Error("async execution timed out for playbook ", playbook.ID)
		case exec_details := <-trigger.Executionsch:
			playbook_id := exec_details.PlaybookId
			exec_id := exec_details.ExecutionId
			if playbook_id == playbook.ID {
				msg := gin.H{
					"execution_id": exec_id,
					"payload":      playbook_id,
				}
				context.JSON(http.StatusOK, msg)
				return
			}
		}
	}
}
