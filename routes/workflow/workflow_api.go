package workflow

import (
	"io"
	"net/http"
	cacao "soarca/internal/cacao"
	validator "soarca/internal/validators"

	"github.com/gin-gonic/gin"
)



func SubmitWorkflow(g *gin.Context) {
	

	jsonData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		logger.Error(component)
		g.JSON(http.StatusBadRequest, gin.H{"error": "failed to marshall json on server side"})
	}

	err = validator.Json[cacao.Playbook](jsonData)
	if err != nil {
		logger.Error(component)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	
    }
	g.JSON(http.StatusOK, gin.H{"message": "JSON data is valid"})

}
