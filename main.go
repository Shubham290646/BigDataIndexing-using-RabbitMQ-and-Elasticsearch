package main

import (
	"info7255-bigdata-app/routes"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	r := routes.SetupRouter()

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
