package routes

import (
	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/controller/informer"
	playbook_routes "soarca/pkg/api/playbook"
	reporter "soarca/pkg/api/reporter"
	status "soarca/pkg/api/status"
	swagger "soarca/pkg/api/swagger"
	"soarca/pkg/api/trigger"

	"github.com/gin-contrib/cors"
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

func Reporter(app *gin.Engine, informer informer.IExecutionInformer) error {
	log.Trace("Setting up reporter routes")
	reporter.Routes(app, informer)
	return nil
}

func Api(app *gin.Engine,
	controller decomposer_controller.IController,
	database database.IController,
) error {
	log.Trace("Trying to setup all Routes")
	// gin.SetMode(gin.ReleaseMode)

	trigger_api := trigger.New(controller, database)
	status.Routes(app)
	trigger.Routes(app, trigger_api)

	return nil
}

func Swagger(app *gin.Engine) {
	swagger.Routes(app)
}

func Cors(app *gin.Engine, origins []string) {

	config := cors.DefaultConfig()
	config.AllowOrigins = origins
	app.Use(cors.New(config))
}
