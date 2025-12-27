package main

import (
	"log"

	"github.com/hunderaweke/tg-unwrapped/internal/server/router"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}
