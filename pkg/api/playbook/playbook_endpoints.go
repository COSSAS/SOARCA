package playbook

import (
	"soarca/internal/controller/database"

	"github.com/gin-gonic/gin"
)

// Main Router for the following endpoints:
// GET     /playbook
// POST    /playbook
// GET     /playbook/playbook-id
// PUT     /playbook/playbook-id
// DELETE  /playbook/playbook-id
func Routes(route *gin.Engine, controller database.IController) {
	playbookController := NewPlaybookController(controller)
	playbook := route.Group("/playbook")
	{
		playbook.GET("/", playbookController.getAllPlaybooks)
		playbook.GET("/meta/", playbookController.getAllPlaybookMetas)
		playbook.POST("/", playbookController.submitPlaybook)
		playbook.GET("/:id", playbookController.getPlaybookByID)
		playbook.PUT("/:id", playbookController.updatePlaybookByID)
		playbook.DELETE("/:id", playbookController.deleteByPlaybookID)

	}
}
