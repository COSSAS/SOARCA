package workflow

import (
	workflowRepository "soarca/database/workflow"

	"github.com/gin-gonic/gin"
)

// Main Router for the following endpoints:
// GET     /workflow
// POST    /workflow
// GET     /workflow/workflow-id
// PUT     /workflow/workflow-id
// DELETE  /workflow/workflow-id
func Routes(route *gin.Engine, workflowRepo workflowRepository.IWorkflowRepository) {
	workflowController := NewWorkflowController(workflowRepo)
	workflow := route.Group("/workflow")
	{
		workflow.GET("/", workflowController.getAllWorkflows)
		workflow.GET("/meta/", workflowController.getAllWorkFlowMetas)
		workflow.POST("/", workflowController.submitWorkflow)
		workflow.GET("/:id", workflowController.getWorkflowByID)
		workflow.PUT("/:id", workflowController.updateWorkflowByID)
		workflow.DELETE("/:id", workflowController.deleteWorkflowByID)

	}
}
