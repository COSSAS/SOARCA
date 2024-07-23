package trigger

import (
	"io"
	"net/http"
	"reflect"
	"time"

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
	Executionsch chan decomposer.ExecutionDetails
}

func New(controller decomposer_controller.IController) *TriggerApi {
	instance := TriggerApi{}
	instance.controller = controller
	// Channel to get back execution details
	instance.Executionsch = make(chan decomposer.ExecutionDetails)
	return &instance
}

func (trigger *TriggerApi) Execute(context *gin.Context) {
	// create new decomposer when execute is called
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
