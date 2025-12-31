package main

import (
	"os"

	"github.com/hunderaweke/tg-unwrapped/internal/logger"
	"github.com/hunderaweke/tg-unwrapped/internal/server/router"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if err := router.Run(); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
