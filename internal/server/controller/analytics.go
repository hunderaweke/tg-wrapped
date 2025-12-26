package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hunderaweke/tg-unwrapped/internal/analyzer"
	"github.com/hunderaweke/tg-unwrapped/internal/storage"
)

type AnalyticsRequest struct {
	Username string `json:"username,omitempty"`
}

func AnalyticsHandler(ctx *gin.Context) {
	var anaReq AnalyticsRequest
	if err := ctx.ShouldBind(&anaReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var analytics *analyzer.Analytics
	redisService, err := storage.NewRedis()
	if err != nil {
		log.Fatal(err)
	}
	ok, err := redisService.Get(anaReq.Username, &analytics)
	if err != nil {
		log.Fatalf("error getting: %v", err)
	}
	if !ok {
		a := analyzer.NewAnalyzer()
		analytics, err = a.ProcessAnalytics(anaReq.Username)
		if err != nil {
			log.Fatal(err)
		}
		err = redisService.Set(anaReq.Username, analytics, 48*time.Hour)
		if err != nil {
			log.Fatal(err)
		}
	}
	ctx.JSON(http.StatusOK, analytics)
}
