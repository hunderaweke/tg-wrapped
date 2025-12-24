package router

import (
	"fmt"
	"net/http"
	"os"

	"github.com/hunderaweke/tg-unwrapped/internal/server/controller"
)

func Run() error {
	http.HandleFunc("/health", controller.HealthHandler)
	fmt.Println("Listening on port: ", os.Getenv("SERVER_PORT"))
	return http.ListenAndServe(":"+os.Getenv("SERVER_PORT"), nil)
}
