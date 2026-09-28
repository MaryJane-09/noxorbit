package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MaryJane-09/noxorbit/backend/config"
	"github.com/MaryJane-09/noxorbit/backend/internal/db"
	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/MaryJane-09/noxorbit/backend/internal/router"
)

func main() {

	fmt.Println("Nexus server running on port", config.AppConfig.ServerPort)
	fmt.Println("Gmail address loaded:", config.AppConfig.GmailAddress != "")
	fmt.Println("App password loaded:", config.AppConfig.GmailAppPassword != "")
	fmt.Println("Database_URL loaded:", config.AppConfig.DatabaseURL != "")
	fmt.Println(otp.Generate(6))
	pool, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	server := router.New(pool)
	err = http.ListenAndServe(config.AppConfig.ServerPort, server)
	if err != nil {
		fmt.Println("Failed to start HTTP server: ", err)
	}
}
