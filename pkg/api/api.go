package api

import (
	"reflect"
	open_api "soarca/api"
	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/controller/informer"
	"soarca/internal/logger"
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

// FinPublic sets up the Fin-protocol endpoints that authenticate via their
// own registration_token/fin_token scheme (register/poll/jobs/status/
// unregister), not SOARCA's admin JWT auth. The caller MUST register these
// before installing the global soarca_admin auth middleware (see
// intializeAuthenticationMiddleware in internal/controller/controller.go) -
// otherwise every Fin call would also require a valid JWT, which a Fin
// process has no way to obtain.
func FinPublic(app *gin.Engine, finHandler *fin_handler.FinHandler) {
	log.Trace("Setting up fin protocol routes (registered ahead of the admin auth middleware - see FinPublic doc comment)")
	FinPublicRoutes(app, finHandler)
}

// FinAdmin sets up the read-only Fin discovery endpoints (list/get). Unlike
// FinPublic, these are ordinary admin/dashboard reads and are expected to
// sit behind the same soarca_admin JWT gate as the rest of the admin API -
// register these the same way/place as routes.Api/routes.Manual/etc.
func FinAdmin(app *gin.Engine, finHandler *fin_handler.FinHandler) {
	log.Trace("Setting up fin discovery routes")
	FinAdminRoutes(app, finHandler)
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
		manualRoutes.GET(":exec_id/:step_execution_id", manualHandler.GetPendingCommand)
		manualRoutes.PUT(":exec_id/:step_execution_id", manualHandler.PutContinue)
	}
}

// FinPublicRoutes registers the Fin-protocol endpoints that authenticate
// via their own registration_token/fin_token scheme, not SOARCA's admin
// JWT auth (see FinPublic's doc comment for why these must be registered
// before the global admin auth middleware is installed):
// POST    /fin/register                (registration-token gated)
// POST    /fin/poll                     (fin-token gated)
// PUT     /fin/jobs/:job_id             (fin-token gated)
// PATCH   /fin/jobs/:job_id/status      (fin-token gated)
// DELETE  /fin/                        (fin-token gated; unregisters the calling fin itself, inferred from the token)
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

// FinAdminRoutes registers the read-only Fin discovery endpoints and the
// admin-initiated delete (ordinary admin/dashboard actions, not
// fin-authenticated):
// GET     /fin/                        (admin/dashboard read)
// GET     /fin/:fin_id                  (admin/dashboard read)
// DELETE  /fin/:fin_id                  (admin/dashboard action; forcibly removes any fin's registration)
func FinAdminRoutes(route *gin.Engine, finHandler *fin_handler.FinHandler) {
	finRoutes := route.Group("/fin")
	{
		finRoutes.GET("/", finHandler.List)
		finRoutes.GET(":fin_id", finHandler.Get)
		finRoutes.DELETE(":fin_id", finHandler.Delete)
	}
}
