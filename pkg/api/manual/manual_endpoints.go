package manual

import (
	"reflect"
	"soarca/internal/logger"

	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// TODO:
// The manual API expose general executions-wide information
// Thus, we need a ManualController that implements an IManualController interface
// The API routes will invoke the IManualController interface instance (the ManualController)
// In turn, the manual capability itself will implement, besides the execution capability,
//	also some IManualInteraction interface, which will expose and consume information specific
//	to a manual interaction (the function that the ManualController will invoke on the manual capability,
//	to GET manual/execution-id/step-id, and POST manual/continue).

// A manual command in CACAO is simply the operation:
// 		{ post_message; wait_for_response (returning a result) }
//
// Agent and target for the command itself make little sense.
// Unless an agent is the intended system that does post_message, and wait_for_response.
// But the targets? For the automated execution, there is no need to specify any.
//
// It is always either only the internal API, or the internal API and ONE integration for manual.
// Env variable: can only have one active manual interactor.
//
// In light of this, for hierarchical and distributed playbooks executions (via multiple playbook actions),
// 	there will be ONE manual integration (besides internal API) per every ONE SOARCA instance.

// The controller manages the manual command infromation and status, like a cache. And interfaces any interactor type (e.g. API, integration)

func Routes(route *gin.Engine, manualHandler *ManualHandler) {
	group := route.Group("/manual")
	{
		group.GET("/", manualHandler.GetPendingCommands)
		group.GET("/:executionId/:stepId", manualHandler.GetPendingCommand)
		group.POST("/manual/continue", manualHandler.PostContinue)
	}
}
