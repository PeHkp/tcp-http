package main

import (
	"log"
	"servidor-tcp/server"
)

func main() {
	server := server.InitServer(":3000")

	log.Fatal(server.Start())

}
