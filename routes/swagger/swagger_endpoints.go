package swagger

import (
	"soarca/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Routes(route *gin.Engine) {
	docs.SwaggerInfo.BasePath = "/"
	swagger := route.Group("/swagger")
	{
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
}
