package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(map[string]string{"health": "ok"})
	fmt.Fprint(w, string(data))
}
