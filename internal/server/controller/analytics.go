package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/hunderaweke/tg-unwrapped/internal/analyzer"
	"github.com/hunderaweke/tg-unwrapped/internal/storage"
)

type AnalyticsRequest struct {
	Username string `json:"username,omitempty"`
}

func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	var anaReq AnalyticsRequest
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err := json.Unmarshal(data, &anaReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
	data, err = json.Marshal(&analytics)
	fmt.Fprint(w, string(data))
}
