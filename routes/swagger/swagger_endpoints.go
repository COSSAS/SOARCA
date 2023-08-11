package swagger

import (
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerfiles "github.com/swaggo/files"
)

func Routes(route *gin.Engine){
	swagger := route.Group("/swagger")
	{
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
}

