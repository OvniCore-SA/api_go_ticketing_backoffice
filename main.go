package main

import (
	"log"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api"
)

func main() {
	if err := api.StartApp(); err != nil {
		log.Fatalf("Error starting Backoffice API: %v", err)
	}
}
