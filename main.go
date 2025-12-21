package main

import (
	"log"

	"github.com/hunderaweke/tg-unwrapped/internal/analyzer"
)

func main() {
	a := analyzer.NewAnalyzer()
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
