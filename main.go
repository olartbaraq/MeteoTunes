package main

import (
	"log"

	"github.com/olartbaraq/meteotunes/api"
)

func main() {
	log.Print("Listening...")
	server := api.NewServer(".")
	server.Start()
}
