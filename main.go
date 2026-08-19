package main

import (
	"log"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	api.SetupApp()
}
