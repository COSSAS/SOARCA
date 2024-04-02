package trigger

import (
	"io"
	"net/http"
	"reflect"

	"soarca/internal/controller/decomposer_controller"
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
	controller decomposer_controller.IController
}

func New(controller decomposer_controller.IController) *TriggerApi {
	instance := TriggerApi{}
	instance.controller = controller
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
	executionDetail, errDecomposer := decomposer.Execute(*playbook)
	if errDecomposer != nil {
		error.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to decode playbook",
			"POST /trigger/playbook",
			executionDetail.ExecutionId.String())
	} else {
		msg := gin.H{
			"execution_id": executionDetail.ExecutionId.String(),
			"payload":      executionDetail.PlaybookId,
		}
		context.JSON(http.StatusOK, msg)
	}
}
