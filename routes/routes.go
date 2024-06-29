package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pmas98/go-auth-service/controllers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.GET("/health", controllers.HealthCheck)
		api.POST("/register", controllers.HandleSignup)
		api.POST("/login", controllers.HandleLogin)
		api.DELETE("/all", controllers.DeleteAll)
	}

	return r
}
