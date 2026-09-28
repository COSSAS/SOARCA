package api

import (
	"reflect"
	open_api "soarca/api"
	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/controller/informer"
	"soarca/internal/logger"
	keymanagement_handler "soarca/pkg/api/keymanagement"
	playbook_handler "soarca/pkg/api/playbook"
	reporter_handler "soarca/pkg/api/reporter"
	status_handler "soarca/pkg/api/status"
	"soarca/pkg/core/capability/manual/interaction"
	"soarca/pkg/keymanagement"

	manual_handler "soarca/pkg/api/manual"

	trigger_handler "soarca/pkg/api/trigger"

	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

func Database(app *gin.Engine,
	controller database.IController,
) error {
	log.Trace("Setting up playbook routes")
	PlaybookRoutes(app, controller)
	return nil
}

func Logging(app *gin.Engine) {
	// app.Use(middelware.LoggingMiddleware(log.Logger))
}

func Reporter(app *gin.Engine, informer informer.IExecutionInformer) error {
	log.Trace("Setting up reporter routes")
	ReporterRoutes(app, informer)
	return nil
}

func Manual(app *gin.Engine, interaction interaction.IInteractionStorage) {
	log.Trace("Setting up manual routes")
	manualHandler := manual_handler.NewManualHandler(interaction)
	ManualRoutes(app, manualHandler)
}
func KeyManagement(app *gin.Engine, key_manager *keymanagement.KeyManagement) {
	log.Trace("Setting up key management routes")
	keyManagement := keymanagement_handler.NewKeyManagementHandler(key_manager)
	KeyManagementRoutes(app, keyManagement)
}

func Api(app *gin.Engine,
	controller decomposer_controller.IController,
	database database.IController,
) error {
	log.Trace("Trying to setup all Routes")
	// gin.SetMode(gin.ReleaseMode)
	triggerHandler := trigger_handler.NewTriggerHandler(controller, database)
	TriggerRoutes(app, triggerHandler)
	StatusRoutes(app)

	return nil
}

func Cors(app *gin.Engine, origins []string) {
	config := cors.DefaultConfig()
	config.AllowOrigins = origins
	app.Use(cors.New(config))
}

func Swagger(app *gin.Engine) {
	swaggerRoutes(app)
}

func swaggerRoutes(route *gin.Engine) {
	open_api.SwaggerInfo.BasePath = "/"
	swaggerRoutes := route.Group("/swagger")
	{
		swaggerRoutes.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
}

// Main Router for the following endpoints:
// GET     /playbook
// POST    /playbook
// GET     /playbook/playbook-id
// PUT     /playbook/playbook-id
// DELETE  /playbook/playbook-id
func PlaybookRoutes(route *gin.Engine, controller database.IController) {
	playbookHandler := playbook_handler.NewPlaybookHandler(controller)
	playbookRoutes := route.Group("/playbook")
	{
		playbookRoutes.GET("/", playbookHandler.GetAllPlaybooks)
		playbookRoutes.POST("/", playbookHandler.SubmitPlaybook)
		playbookRoutes.GET("/meta/", playbookHandler.GetAllPlaybookMetas)
		playbookRoutes.GET("/:id", playbookHandler.GetPlaybookByID)
		playbookRoutes.PUT("/:id", playbookHandler.UpdatePlaybookByID)
		playbookRoutes.DELETE("/:id", playbookHandler.DeleteByPlaybookID)

	}
}

// Main Router for the following endpoints:
// GET     /reporter
// GET     /reporter/{execution-id}
func ReporterRoutes(route *gin.Engine, informer informer.IExecutionInformer) {
	reportHandler := reporter_handler.NewReportHandler(informer)
	reportRoutes := route.Group("/reporter")
	{
		reportRoutes.GET("/", reportHandler.GetExecutions)
		reportRoutes.GET("/:id", reportHandler.GetExecutionReport)
	}
}

// GET     /status
// GET     /status/ping
func StatusRoutes(route *gin.Engine) {
	router := route.Group("/status")
	{
		router.GET("/", status_handler.GetApi)
		router.GET("/ping", status_handler.GetPong)

	}
}

func TriggerRoutes(route *gin.Engine, triggerHandler *trigger_handler.TriggerHandler) {
	triggerRoutes := route.Group("/trigger")
	{
		triggerRoutes.POST("/playbook", triggerHandler.Execute)
		triggerRoutes.POST("/playbook/:id", triggerHandler.ExecuteById)
	}
}

func ManualRoutes(route *gin.Engine, manualHandler *manual_handler.ManualHandler) {
	manualRoutes := route.Group("/manual")
	{
		manualRoutes.GET("/", manualHandler.GetPendingCommands)
		manualRoutes.GET(":exec_id/:step_id", manualHandler.GetPendingCommand)
		manualRoutes.POST("/continue", manualHandler.PostContinue)
	}
}

func KeyManagementRoutes(route *gin.Engine, keyManagementHandler *keymanagement_handler.KeyManagementHandler) {
	keyManagementRoutes := route.Group("/keymanagement")
	{
		keyManagementRoutes.GET("/", keyManagementHandler.GetKeys)
		keyManagementRoutes.PUT("/:keyname", keyManagementHandler.AddKey)
		keyManagementRoutes.PATCH("/:keyname", keyManagementHandler.UpdateKey)
		keyManagementRoutes.DELETE("/:keyname", keyManagementHandler.RevokeKey)
	}
}
