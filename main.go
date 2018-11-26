package main

import (
	"log"
	"os"

	"github.com/ryanrolds/pedantic_orderliness/site"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}

	site := site.NewSite(port, env)
	err := site.Run()
	if err != nil {
		log.Panic(err)
	}

	os.Exit(0)
}
