package trigger

import (
	"io"
	"net/http"
	"reflect"
	"time"

	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/decomposer"
	"soarca/logger"
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

// trigger
//
//	@Summary	trigger a playbook by id that is stored in SOARCA
//	@Schemes
//	@Description	trigger playbook by id
//	@Tags			trigger
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"playbook ID"
//	@Success		200			{object}	api.Execution
//	@failure		400			{object}	api.Error
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

	// create new decomposer when execute is called
	decomposer := trigger.controller.NewDecomposer()
	executionDetail, errDecomposer := decomposer.Execute(playbook)
	if errDecomposer != nil {
		error.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to decode playbook",
			"POST /trigger/playbook/"+id,
			executionDetail.ExecutionId.String())
	} else {
		msg := gin.H{
			"execution_id": executionDetail.ExecutionId.String(),
			"payload":      executionDetail.PlaybookId,
		}
		context.JSON(http.StatusOK, msg)
	}

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
	decomposer := trigger.controller.NewDecomposer()
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

	go decomposer.ExecuteAsync(*playbook, trigger.Executionsch)

	// Hard coding the timer to return execution id
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
