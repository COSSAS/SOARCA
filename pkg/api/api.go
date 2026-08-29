package api

import (
	"reflect"
	open_api "soarca/api"
	"soarca/internal/controller/database"
	"soarca/internal/controller/informer"
	"soarca/internal/logger"
	"soarca/internal/services"
	"soarca/internal/services/playbook"
	"soarca/internal/services/reporter"
	triggerservice "soarca/internal/services/trigger"
	playbook_handler "soarca/pkg/api/playbook"
	reporter_handler "soarca/pkg/api/reporter"
	status_handler "soarca/pkg/api/status"
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

// ============================================================================
// Handler constructors (for callers that inject services directly)
// ============================================================================

func NewTriggerHandler(svc services.TriggerService) *trigger_handler.TriggerHandler {
	return trigger_handler.NewTriggerHandler(svc)
}

func NewManualHandler(inbox services.ManualInbox) *manual_handler.ManualHandler {
	return manual_handler.NewManualHandler(inbox)
}

// ============================================================================
// Route registration functions
// ============================================================================

func Cors(app *gin.Engine, origins []string) {
	config := cors.DefaultConfig()
	config.AllowOrigins = origins
	app.Use(cors.New(config))
}

func Logging(app *gin.Engine) {}

func Swagger(app *gin.Engine) { swaggerRoutes(app) }

func swaggerRoutes(route *gin.Engine) {
	open_api.SwaggerInfo.BasePath = "/"
	swaggerRoutes := route.Group("/swagger")
	{
		swaggerRoutes.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
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

// PlaybookRoutesWithService registers playbook CRUD routes using an injected service.
func PlaybookRoutesWithService(route *gin.Engine, svc services.PlaybookService) {
	log.Trace("Setting up playbook routes")
	playbookHandler := playbook_handler.NewPlaybookHandler(svc)
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

// ReporterRoutesWithService registers reporter routes using an injected service.
func ReporterRoutesWithService(route *gin.Engine, svc services.ReporterService) {
	log.Trace("Setting up reporter routes")
	reportHandler := reporter_handler.NewReportHandler(svc)
	reportRoutes := route.Group("/reporter")
	{
		reportRoutes.GET("/", reportHandler.GetExecutions)
		reportRoutes.GET("/:id", reportHandler.GetExecutionReport)
	}
}

func FinPublic(app *gin.Engine, finHandler *fin_handler.FinHandler) {
	log.Trace("Setting up fin protocol routes (registered ahead of the admin auth middleware - see FinPublic doc comment)")
	FinPublicRoutes(app, finHandler)
}

func FinAdmin(app *gin.Engine, finHandler *fin_handler.FinHandler) {
	log.Trace("Setting up fin discovery routes")
	FinAdminRoutes(app, finHandler)
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

// ============================================================================
// Legacy helpers — kept for tests that still use the old controller-based API
// ============================================================================

func Database(app *gin.Engine, controller database.IController) error {
	log.Trace("Setting up playbook routes")
	PlaybookRoutes(app, controller)
	return nil
}

func Reporter(app *gin.Engine, inf informer.IExecutionInformer) error {
	log.Trace("Setting up reporter routes")
	ReporterRoutes(app, inf)
	return nil
}

func Api(app *gin.Engine, executionRuntime services.ExecutionRuntime, db database.IController) error {
	log.Trace("Setting up trigger routes")
	triggerHandler := trigger_handler.NewTriggerHandler(triggerservice.New(executionRuntime, db.GetPlaybookStore()))
	TriggerRoutes(app, triggerHandler)
	StatusRoutes(app)
	return nil
}

func PlaybookRoutes(route *gin.Engine, controller database.IController) {
	playbookHandler := playbook_handler.NewPlaybookHandler(playbook.New(controller.GetPlaybookStore()))
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

func ReporterRoutes(route *gin.Engine, inf informer.IExecutionInformer) {
	reportHandler := reporter_handler.NewReportHandler(reporter.New(inf))
	reportRoutes := route.Group("/reporter")
	{
		reportRoutes.GET("/", reportHandler.GetExecutions)
		reportRoutes.GET("/:id", reportHandler.GetExecutionReport)
	}
}

