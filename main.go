package main

import (
	"os"

	"github.com/ryanrolds/pedantic_orderliness/site"
	"github.com/sirupsen/logrus"
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}

	log := setupLog(env)

	hostname, err := os.Hostname()
	if err != nil {
		log.WithError(err).Error("Problem getting hostname")
		hostname = "unknown"
	}

	log = log.WithFields(logrus.Fields{
		"env":  env,
		"host": hostname,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	site := site.NewSite(port, env, log)
	err = site.Run()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func setupLog(env string) *logrus.Entry {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	return logrus.NewEntry(log)
}
