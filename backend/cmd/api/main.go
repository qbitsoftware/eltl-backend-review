package main

import (
	"log"
	"table-tennis/internal/app/server"
)

func main() {
	//env file reading

	//start server here
	log.Fatal(server.Start())
}
