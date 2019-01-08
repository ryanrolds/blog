package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ryanrolds/pedantic_orderliness/site"
	log "github.com/sirupsen/logrus"
)

var contentDir = "./content/"
var wait = time.Second * 30

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "test"
	}

	// Prepare site and get handler
	site := site.NewSite(port, env, contentDir)
	handler, err := site.GetHandler()
	if err != nil {
		log.Fatal(err)
	}

	// Prepare server
	server := http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Run our server
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// Wait on SIGINT
	signal.Notify(c, os.Interrupt)

	// Block until we get the signal
	<-c

	// Create timeout in case shutdown runs long
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Shutdown
	log.Println("Shutting down")
	server.Shutdown(ctx)

	os.Exit(0)
}
