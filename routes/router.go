package routes

import (
	middelware "soarca/middelware"
	coa_routes "soarca/routes/coa"
	operator "soarca/routes/operator"
	status "soarca/routes/status"
	swagger "soarca/routes/swagger"
	workflow_routes "soarca/routes/workflow"

	gin "github.com/gin-gonic/gin"
)

// POST    /operator/coa/coa-id

func Setup() *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(middelware.LoggingMiddleware(logger))
	coa_routes.Routes(app)
	workflow_routes.Routes(app)
	status.Routes(app)
	operator.Routes(app)
	swagger.Routes(app)

	return app
}


