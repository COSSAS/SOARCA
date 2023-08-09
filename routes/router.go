package routes

import (
	coa_routes "soarca/routes/coa"
	operator "soarca/routes/operator"
	status "soarca/routes/status"
	workflow_routes "soarca/routes/workflow"

	gin "github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// POST    /operator/coa/coa-id

func Setup() *gin.Engine {
	app := gin.New()
	// f, _ := os.Create("log/api.log")
	//gin.DisableConsoleColor()
	//gin.DefaultWriter = io.MultiWriter(f)
	
	// app.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	// 	return fmt.Sprintf("%s - - [%s] \"%s %s %s %d %s \" \" %s\" \" %s\"\n",
	// 		param.ClientIP,
	// 		param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
	// 		param.Method,
	// 		param.Path,
	// 		param.Request.Proto,
	// 		param.StatusCode,
	// 		param.Latency,
	// 		param.Request.UserAgent(),
	// 		param.ErrorMessage,
	// 	)
	// }))

	coa_routes.Routes(app)
	workflow_routes.Routes(app)
	status.Routes(app)
	operator.Routes(app)
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return app
}


