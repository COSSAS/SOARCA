package reporter

import (
	"soarca/internal/controller/informer"

	"github.com/gin-gonic/gin"
)

// Main Router for the following endpoints:
// GET     /reporter
// GET     /reporter/{execution-id}
func Routes(route *gin.Engine, informer informer.IExecutionInformer) {
	executionInformer := NewExecutionInformer(informer)
	report := route.Group("/reporter")
	{
		report.GET("", executionInformer.getExecutions)
		report.GET("/:id", executionInformer.getExecutionReport)
	}
}
