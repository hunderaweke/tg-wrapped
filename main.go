package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hunderaweke/tg-unwrapped/internal/analyzer"
	"github.com/hunderaweke/tg-unwrapped/internal/storage"
)

func main() {
	var analytics *analyzer.Analytics
	channelUsername := "dagmawi_babi"
	redisService, err := storage.NewRedis()
	if err != nil {
		log.Fatal(err)
	}
	ok, err := redisService.Get(channelUsername, &analytics)
	if err != nil {
		log.Fatalf("error getting: %v", err)
	}
	if ok {
		fmt.Printf("%+v\n", analytics)
		return
	}
	a := analyzer.NewAnalyzer()
	analytics, err = a.ProcessAnalytics(channelUsername)
	if err != nil {
		log.Fatal(err)
	}
	err = redisService.Set(channelUsername, analytics, 48*time.Hour)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", analytics)
}
