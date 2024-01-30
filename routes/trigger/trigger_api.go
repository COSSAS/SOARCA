package trigger

import (
	"io"
	"net/http"
	"reflect"
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
	decomposer decomposer.IDecomposer
}

func New(decomposer decomposer.IDecomposer) *TriggerApi {
	var instance = TriggerApi{}
	instance.decomposer = decomposer
	return &instance
}

func (trigger *TriggerApi) Execute(context *gin.Context) {
	jsonData, errIo := io.ReadAll(context.Request.Body)
	if errIo != nil {
		log.Error("failed")
		error.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to marshall json on server side",
			"POST /trigger/workflow", "")
		return
	}
	// playbook := cacao.Decode(jsonData)
	playbook := decoder.DecodeValidate(jsonData)
	if playbook == nil {
		error.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to decode playbook",
			"POST /trigger/workflow", "")
		return
	}
	executionDetail, errDecomposer := trigger.decomposer.Execute(*playbook)
	if errDecomposer != nil {
		error.SendErrorResponse(context, http.StatusBadRequest,
			"Failed to decode playbook",
			"POST /trigger/workflow",
			executionDetail.ExecutionId.String())
	} else {
		msg := gin.H{
			"execution_id": executionDetail.ExecutionId.String(),
			"payload":      executionDetail.PlaybookId,
		}
		context.JSON(http.StatusOK, msg)
	}
}
