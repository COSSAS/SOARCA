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
func SetupRoutes(workflowRepo *workflowRepository.WorkflowRepository,
	decomposer decomposer.IDecomposer) *gin.Engine {
	log.Trace("Trying to setup all Routes")
	// gin.SetMode(gin.ReleaseMode)

	trigger_api := trigger.New(decomposer)

	app := gin.New()
	// app.Use(middelware.LoggingMiddleware(log.Logger))
	coa_routes.Routes(app)
	workflow_routes.Routes(app, workflowRepo)
	status.Routes(app)
	operator.Routes(app)
	swagger.Routes(app)
	trigger.Routes(app, trigger_api)

	return app
}
