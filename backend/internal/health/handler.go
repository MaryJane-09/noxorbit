package health

import (
	"encoding/json"
	"log"
	"net/http"
)

type Data struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Version string `json:"version"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := Data{
		Status:  "success",
		Message: "NoxOrbit API is running",
		Version: "v1",
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Encode health response:", err)
	}
}
