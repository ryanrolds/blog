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

  site := &site.Site{
    Port: port,
  }

  err := site.Run()
  if err != nil {
    log.Panic(err)
  }

  os.Exit(0)
}
