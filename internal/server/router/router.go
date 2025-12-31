package router

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hunderaweke/tg-unwrapped/internal/logger"
	"github.com/hunderaweke/tg-unwrapped/internal/server/controller"
	"github.com/hunderaweke/tg-unwrapped/internal/storage"
)

func Run() error {
	// Initialize logger based on environment
	env := os.Getenv("ENV")
	if env == "production" {
		logger.Init(slog.LevelInfo, true)
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger.Init(slog.LevelDebug, false)
	}

	logger.Info("Starting TG-Wrapped server", "env", env)

	redisService, err := storage.NewRedis()
	if err != nil {
		logger.Error("Failed to initialize Redis", "error", err)
		return err
	}

	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "channel-profiles"
		logger.Warn("MINIO_BUCKET not set, using default", "default", bucket)
	}

	minioClient, err := storage.NewMinioBucket(bucket)
	if err != nil {
		logger.Error("Failed to initialize Minio", "error", err)
		return err
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(cors.Default())

	// Set trusted proxies
	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Warn("Failed to set trusted proxies", "error", err)
	}

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
			logger.Error("Failed to generate access URL", "object", objectName, "error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.Redirect(http.StatusTemporaryRedirect, url)
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "7000"
		logger.Warn("SERVER_PORT not set, using default", "default", port)
	}

	logger.Info("Server starting", "port", port)
	return router.Run(":" + port)
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		logger.Info("HTTP request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"duration", time.Since(start),
			"client_ip", c.ClientIP())
	}
}
