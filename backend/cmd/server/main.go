package main

import (
	"log"
	"net/http"

	"github.com/MaryJane-09/noxorbit/backend/config"
	"github.com/MaryJane-09/noxorbit/backend/internal/db"
	"github.com/MaryJane-09/noxorbit/backend/internal/router"
)

func main() {
	pool, err := db.Connect()
	if err != nil {
		log.Fatal("Connect to database: ", err)
	}
	defer pool.Close()

	server := router.New(pool)
	log.Println("NoxOrbit server starting on", config.AppConfig.ServerPort)
	// A failed listener or unexpected server stop must terminate the process
	// with a nonzero exit status instead of only printing an error.
	log.Fatal(http.ListenAndServe(config.AppConfig.ServerPort, server))
}
