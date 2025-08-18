package routes

import (
	controller "github.com/0xk4n3ki/restaurant-management-backend/controllers"
	"github.com/0xk4n3ki/restaurant-management-backend/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/users/signup", controller.SignUp())
	incomingRoutes.POST("/users/login", controller.Login())
	incomingRoutes.POST("/users/refresh", controller.RefreshToken())

	userGroup := incomingRoutes.Group("/users")
	userGroup.Use(middleware.Authentication())
	{
		incomingRoutes.GET("/", controller.GetUsers())
		incomingRoutes.GET("/:user_id", controller.GetUser())
	}
}
