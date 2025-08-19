package main

import (
	"log"
	"net/http"
	"os"

	"github.com/0xk4n3ki/restaurant-management-backend/middleware"
	"github.com/0xk4n3ki/restaurant-management-backend/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.UserRoutes(router)

	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.Use(middleware.Authentication())

	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	err := router.Run(":" + port)
	if err != nil {
		log.Panic("failed to run the router")
	}
}
