package main

import (
	"log"
	"net/http"
	"time"

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

	server := &http.Server{
		Addr:              config.AppConfig.ServerPort,
		Handler:           router.New(pool),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Println("NoxOrbit server starting on", config.AppConfig.ServerPort)
	log.Fatal(server.ListenAndServe())
}
