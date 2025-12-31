package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hunderaweke/tg-unwrapped/internal/analyzer"
	"github.com/hunderaweke/tg-unwrapped/internal/logger"
	"github.com/hunderaweke/tg-unwrapped/internal/storage"
)

type AnalyticsRequest struct {
	Username string `json:"username,omitempty"`
}

func AnalyticsHandler(redisService *storage.RedisService, minioClient *storage.MinioClient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log := logger.With("handler", "AnalyticsHandler")

		var anaReq AnalyticsRequest
		if err := ctx.ShouldBindJSON(&anaReq); err != nil {
			log.Warn("Invalid request body", "error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if anaReq.Username == "" {
			log.Warn("Username is required")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
			return
		}

		log = logger.With("handler", "AnalyticsHandler", "username", anaReq.Username)
		log.Info("Processing analytics request")

		var analytics *analyzer.Analytics
		ok, err := redisService.Get(anaReq.Username, &analytics)
		if err != nil {
			log.Warn("Failed to get from cache, proceeding without cache", "error", err)
			// Continue without cache, don't fail
		}

		if ok && analytics != nil {
			log.Info("Returning cached analytics")
			ctx.JSON(http.StatusOK, analytics)
			return
		}

		a, err := analyzer.NewAnalyzer(minioClient)
		if err != nil {
			log.Error("Failed to create analyzer", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to initialize analyzer",
				"details": err.Error(),
			})
			return
		}

		analytics, err = a.ProcessAnalytics(anaReq.Username)
		if err != nil {
			log.Error("Failed to process analytics", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to process analytics",
				"details": err.Error(),
			})
			return
		}

		// Cache the result (non-fatal if fails)
		if err := redisService.Set(anaReq.Username, analytics, 48*time.Hour); err != nil {
			log.Warn("Failed to cache analytics result", "error", err)
		}

		log.Info("Analytics processed successfully")
		ctx.JSON(http.StatusOK, analytics)
	}
}
