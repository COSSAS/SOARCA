package routes

import (
	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer"
	coa_routes "soarca/routes/coa"
	operator "soarca/routes/operator"
	playbook_routes "soarca/routes/playbook"
	status "soarca/routes/status"
	swagger "soarca/routes/swagger"
	"soarca/routes/trigger"

	gin "github.com/gin-gonic/gin"
)

// POST    /operator/coa/coa-id

// Function setup the required routes for the API layout.
// Requires database dependency injection.

func Database(app *gin.Engine,
	controller database.IController,
) error {
	playbook_routes.Routes(app, controller)
	return nil
}

func Logging(app *gin.Engine) {
	// app.Use(middelware.LoggingMiddleware(log.Logger))
}

func Api(app *gin.Engine,
	controller decomposer.IController,
) error {
	log.Trace("Trying to setup all Routes")
	// gin.SetMode(gin.ReleaseMode)

	trigger_api := trigger.New(controller)

	coa_routes.Routes(app)

	status.Routes(app)
	operator.Routes(app)
	trigger.Routes(app, trigger_api)

	return nil
}

func Swagger(app *gin.Engine) {
	swagger.Routes(app)
}
