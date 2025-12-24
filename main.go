package main

import (
	"log"

	"github.com/hunderaweke/tg-unwrapped/internal/server/router"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// var analytics *analyzer.Analytics
	// channelUsername := "dagmawi_babi"
	// redisService, err := storage.NewRedis()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// ok, err := redisService.Get(channelUsername, &analytics)
	// if err != nil {
	// 	log.Fatalf("error getting: %v", err)
	// }
	// if ok {
	// 	fmt.Printf("%+v\n", analytics)
	// 	return
	// }
	// a := analyzer.NewAnalyzer()
	// analytics, err = a.ProcessAnalytics(channelUsername)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = redisService.Set(channelUsername, analytics, 48*time.Hour)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%+v\n", analytics)
	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}
