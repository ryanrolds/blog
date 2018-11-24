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

	site, err := site.NewSite(port, env)
	if err != nil {
		log.Panic(err)
	}

	err = site.Run()
	if err != nil {
		log.Panic(err)
	}

	os.Exit(0)
}
