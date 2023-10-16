package routes

import (
	gin "github.com/gin-gonic/gin"
	workflowRepository "soarca/database/workflow"
	coa_routes "soarca/routes/coa"
	operator "soarca/routes/operator"
	status "soarca/routes/status"
	swagger "soarca/routes/swagger"
	workflow_routes "soarca/routes/workflow"
)

// POST    /operator/coa/coa-id

// Function setup the required routes for the API layout.
// Requires database dependency injection.
func SetupRoutes(workflowRepo *workflowRepository.WorkflowRepository) *gin.Engine {
	log.Trace("Trying to setup all Routes")
	// gin.SetMode(gin.ReleaseMode)

	app := gin.New()
	// app.Use(middelware.LoggingMiddleware(log.Logger))
	coa_routes.Routes(app)
	workflow_routes.Routes(app, workflowRepo)
	status.Routes(app)
	operator.Routes(app)
	swagger.Routes(app)

	return app
}
