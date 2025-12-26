package router

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hunderaweke/tg-unwrapped/internal/server/controller"
)

func Run() error {
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/health", controller.HealthHandler)
	router.POST("/analytics", controller.AnalyticsHandler)
	return router.Run(":" + os.Getenv("SERVER_PORT"))
}
