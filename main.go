package main

import (
	"os"

	"github.com/ryanrolds/pedantic_orderliness/site"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})

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
		log.Fatal(err)
	}

	os.Exit(0)
}
