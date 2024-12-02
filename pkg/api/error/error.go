package error

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func SendErrorResponse(g *gin.Context,
	status int,
	message string,
	orginal_call string,
	downstream string) {
	msg := gin.H{
		"status":          strconv.Itoa(status),
		"message":         message,
		"original-call":   orginal_call,
		"downstream-call": downstream,
	}
	g.JSON(status, msg)
}
