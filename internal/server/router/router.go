package router

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hunderaweke/tg-unwrapped/internal/server/controller"
	"github.com/hunderaweke/tg-unwrapped/internal/storage"
)

func Run() error {
	redisService, err := storage.NewRedis()
	if err != nil {
		return err
	}
	minioClient, err := storage.NewMinioBucket(os.Getenv("MINIO_BUCKET"))
	if err != nil {
		return err
	}
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/health", controller.HealthHandler)
	router.POST("/analytics", controller.AnalyticsHandler(redisService, minioClient))
	router.GET("/profiles/:objectName", func(ctx *gin.Context) {
		objectName := ctx.Param("objectName")
		if objectName == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "the name of the file is required"})
			return
		}
		url, err := minioClient.GenerateAccessURL(objectName, 48*time.Hour)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.Redirect(http.StatusTemporaryRedirect, url)
	})
	return router.Run(":" + os.Getenv("SERVER_PORT"))
}
