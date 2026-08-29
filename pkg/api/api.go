package api

import (
	"reflect"
	open_api "soarca/api"
	"soarca/internal/controller/database"
	"soarca/internal/controller/informer"
	"soarca/internal/logger"
	"soarca/internal/services"
	playbook_handler "soarca/pkg/api/playbook"
	reporter_handler "soarca/pkg/api/reporter"
	status_handler "soarca/pkg/api/status"
	"soarca/pkg/core/capability/manual/interaction"

	manual_handler "soarca/pkg/api/manual"

	fin_handler "soarca/pkg/api/fin"

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

func Database(app *gin.Engine, controller database.IController) error {
	log.Trace("Setting up playbook routes")
	PlaybookRoutes(app, controller)
	return nil
}

func Logging(app *gin.Engine) {}

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

func FinPublic(app *gin.Engine, finHandler *fin_handler.FinHandler) {
	log.Trace("Setting up fin protocol routes (registered ahead of the admin auth middleware - see FinPublic doc comment)")
	FinPublicRoutes(app, finHandler)
}

func FinAdmin(app *gin.Engine, finHandler *fin_handler.FinHandler) {
	log.Trace("Setting up fin discovery routes")
	FinAdminRoutes(app, finHandler)
}

func Api(app *gin.Engine, executionRuntime services.ExecutionRuntime, database database.IController) error {
	log.Trace("Trying to setup all Routes")
	triggerHandler := trigger_handler.NewTriggerHandler(executionRuntime, database.GetPlaybookStore())
	TriggerRoutes(app, triggerHandler)
	StatusRoutes(app)
	return nil
}

func Cors(app *gin.Engine, origins []string) {
	config := cors.DefaultConfig()
	config.AllowOrigins = origins
	app.Use(cors.New(config))
}

func Swagger(app *gin.Engine) { swaggerRoutes(app) }

func swaggerRoutes(route *gin.Engine) {
	open_api.SwaggerInfo.BasePath = "/"
	swaggerRoutes := route.Group("/swagger")
	{
		swaggerRoutes.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
}

func PlaybookRoutes(route *gin.Engine, controller database.IController) {
	playbookHandler := playbook_handler.NewPlaybookHandler(controller.GetPlaybookStore())
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

func ReporterRoutes(route *gin.Engine, informer informer.IExecutionInformer) {
	reportHandler := reporter_handler.NewReportHandler(informer)
	reportRoutes := route.Group("/reporter")
	{
		reportRoutes.GET("/", reportHandler.GetExecutions)
		reportRoutes.GET("/:id", reportHandler.GetExecutionReport)
	}
}

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
		manualRoutes.GET(":exec_id/:step_execution_id", manualHandler.GetPendingCommand)
		manualRoutes.PUT(":exec_id/:step_execution_id", manualHandler.PutContinue)
	}
}

func FinPublicRoutes(route *gin.Engine, finHandler *fin_handler.FinHandler) {
	finRoutes := route.Group("/fin")
	{
		finRoutes.POST("/register", finHandler.Register)
		finAuthenticated := finRoutes.Group("")
		finAuthenticated.Use(finHandler.RequireFinToken)
		{
			finAuthenticated.POST("/poll", finHandler.Poll)
			finAuthenticated.PUT("jobs/:job_id", finHandler.SubmitResult)
			finAuthenticated.PATCH("jobs/:job_id/status", finHandler.StatusPing)
			finAuthenticated.DELETE("/", finHandler.Unregister)
		}
	}
}

func FinAdminRoutes(route *gin.Engine, finHandler *fin_handler.FinHandler) {
	finRoutes := route.Group("/fin")
	{
		finRoutes.GET("/", finHandler.List)
		finRoutes.GET(":fin_id", finHandler.Get)
		finRoutes.DELETE(":fin_id", finHandler.Delete)
	}
}
