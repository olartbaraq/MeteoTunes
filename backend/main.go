package main

import (
	"log"
)

func main() {
	log.Print("Listening...")
	server := NewServer(".")
	server.Start()
}
