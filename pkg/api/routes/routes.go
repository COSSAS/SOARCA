package routes

import (
	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/controller/informer"
	playbook_routes "soarca/pkg/api/playbook"
	reporter "soarca/pkg/api/reporter"
	status "soarca/pkg/api/status"
	"soarca/pkg/api/trigger"

	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
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

func Cors(app *gin.Engine, origins []string) {
	config := cors.DefaultConfig()
	config.AllowOrigins = origins
	app.Use(cors.New(config))
}

func SwaggerRoutes(route *gin.Engine) {
	api.SwaggerInfo.BasePath = "/"
	swagger := route.Group("/swagger")
	{
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
}

// Main Router for the following endpoints:
// GET     /playbook
// POST    /playbook
// GET     /playbook/playbook-id
// PUT     /playbook/playbook-id
// DELETE  /playbook/playbook-id
func PlaybookRoutes(route *gin.Engine, controller database.IController) {
	playbookHandler := NewPlaybookHandler(controller)
	playbook := route.Group("/playbook")
	{
		playbook.GET("/", playbookHandler.GetAllPlaybooks)
		playbook.GET("/meta/", playbookHandler.GetAllPlaybookMetas)
		playbook.POST("/", playbookHandler.SubmitPlaybook)
		playbook.GET("/:id", playbookHandler.GetPlaybookByID)
		playbook.PUT("/:id", playbookHandler.UpdatePlaybookByID)
		playbook.DELETE("/:id", playbookHandler.DeleteByPlaybookID)

	}
}

// Main Router for the following endpoints:
// GET     /reporter
// GET     /reporter/{execution-id}
func ReporterRoutes(route *gin.Engine, informer informer.IExecutionInformer) {
	executionInformer := NewExecutionInformer(informer)
	report := route.Group("/reporter")
	{
		report.GET("/", executionInformer.getExecutions)
		report.GET("/:id", executionInformer.getExecutionReport)
	}
}

// GET     /status
// GET     /status/ping
func StepRoutes(route *gin.Engine) {
	router := route.Group("/status")
	{
		router.GET("/", Api)
		router.GET("/ping", Pong)

	}
}

func TriggerRoutes(route *gin.Engine, trigger *TriggerApi) {
	group := route.Group("/trigger")
	{
		group.POST("/playbook", trigger.Execute)
		group.POST("/playbook/:id", trigger.ExecuteById)
	}
}
