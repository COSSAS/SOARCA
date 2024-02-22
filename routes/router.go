package routes

import (
	playbookRepository "soarca/database/playbook"
	"soarca/internal/decomposer"
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
	playbookRepo playbookRepository.IPlaybookRepository,
) error {
	playbook_routes.Routes(app, playbookRepo)
	return nil
}

func Logging(app *gin.Engine) {
	// app.Use(middelware.LoggingMiddleware(log.Logger))
}

func Api(app *gin.Engine,
	decomposer decomposer.IDecomposer,
) error {
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
