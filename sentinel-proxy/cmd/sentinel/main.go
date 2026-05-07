package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/omar/sentinel-proxy/internal/proxy"
	"github.com/omar/sentinel-proxy/internal/rules"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// Load WAF rules
	rules.LoadRules()

	app := proxy.NewApp()
	app.Start()
}
