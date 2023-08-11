package loggerfactory

import (
	"time"

	gin "github.com/gin-gonic/gin"
	logrus "github.com/sirupsen/logrus"
)


func LoggingMiddleware(fl *logrus.Logger) gin.HandlerFunc {
  return func(ctx *gin.Context) {

    startTime := time.Now()
    ctx.Next()
    endTime := time.Now()   
    latencyTime := endTime.Sub(startTime)
    reqMethod := ctx.Request.Method
    reqUri := ctx.Request.RequestURI
    statusCode := ctx.Writer.Status()
    clientIP := ctx.ClientIP()

    fl.WithFields(logrus.Fields{
      "METHOD":     reqMethod,
      "URI":        reqUri,
      "STATUS":     statusCode,
      "LATENCY":    latencyTime,
      "CLIENT_IP":  clientIP,
    }).Info("HTTP REQUEST")

    ctx.Next()
  }
}