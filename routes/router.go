package routes

import (
	workflowRepository "soarca/database/workflow"
	"soarca/internal/decomposer"
	coa_routes "soarca/routes/coa"
	operator "soarca/routes/operator"
	status "soarca/routes/status"
	swagger "soarca/routes/swagger"
	"soarca/routes/trigger"
	workflow_routes "soarca/routes/workflow"

	gin "github.com/gin-gonic/gin"
)

// POST    /operator/coa/coa-id

// Function setup the required routes for the API layout.
// Requires database dependency injection.

func Database(app *gin.Engine,
	workflowRepo workflowRepository.IWorkflowRepository) error {
	workflow_routes.Routes(app, workflowRepo)
	return nil
}

func Logging(app *gin.Engine) {
	//app.Use(middelware.LoggingMiddleware(log.Logger))
}

func Api(app *gin.Engine,
	decomposer decomposer.IDecomposer) error {
	log.Trace("Trying to setup all Routes")
	// gin.SetMode(gin.ReleaseMode)

	trigger_api := trigger.New(decomposer)

	coa_routes.Routes(app)

	status.Routes(app)
	operator.Routes(app)
	trigger.Routes(app, trigger_api)

	return nil
}

func Swagger(app *gin.Engine) {
	swagger.Routes(app)
}
