package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/hunderaweke/tg-unwrapped/internal/analyzer"
)

func main() {
	a := analyzer.NewAnalyzer()
	analytics, err := a.ProcessAnalytics("dagmawi_babi")
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(analytics)
	file, err := os.Create("parsed.json")
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
